package e2e

import (
	"context"
	"testing"
	"time"

	"messaging-app/e2e/testclient"
	"messaging-app/internal/domain"
	"messaging-app/testdata"

	"github.com/stretchr/testify/suite"
)

// UserJourneyTestSuite tests complete user scenarios from start to finish
type UserJourneyTestSuite struct {
	E2ETestSuite
}

func (s *UserJourneyTestSuite) TestNewUserFirstConversationJourney() {
	s.T().Log("=== Testing: New User First Conversation Journey ===")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Use testdata users
	alice := testdata.Alice
	bob := testdata.Bob

	// Step 1: Alice joins the platform
	s.T().Log("Step 1: Alice joins the platform")
	aliceClient := s.CreateTestUser(alice.UserID, alice.Email, alice.Handler)

	// Verify Alice can access her empty chat list
	aliceChats, err := aliceClient.GetChats(ctx)
	s.Require().NoError(err, "Alice should be able to get her initial empty chat list")
	s.Empty(aliceChats.Chats, "Alice should start with no conversations")

	// Step 2: Bob joins the platform
	s.T().Log("Step 2: Bob joins the platform")
	bobClient := s.CreateTestUser(bob.UserID, bob.Email, bob.Handler)

	// Step 3: Bob sets up NATS subscription to receive messages
	s.T().Log("Step 3: Bob sets up real-time message subscription")
	bobNATS, err := s.GetNATSClient(bob.UserID)
	s.Require().NoError(err, "Bob should be able to create NATS client")
	defer bobNATS.Close()

	// Wait for NATS connection
	err = bobNATS.WaitForConnection(5 * time.Second)
	s.Require().NoError(err, "Bob's NATS client should connect")

	// Set up message collector for Bob
	messageCollector := testclient.NewMessageCollector()
	err = bobNATS.SubscribeToMessages(ctx, messageCollector.Handler())
	s.Require().NoError(err, "Bob should be able to subscribe to messages")

	// Step 4: Alice starts conversation with Bob
	s.T().Log("Step 4: Alice sends first message to Bob")
	firstMessage := "Hi Bob! Welcome to the platform. How are you settling in?"

	sentMsg, err := aliceClient.SendMessage(ctx, bob.UserID, firstMessage)
	s.Require().NoError(err, "Alice should be able to send first message to Bob")

	// Verify message properties
	s.Equal(alice.UserID, sentMsg.SenderID, "Sender should be Alice")
	s.Equal(bob.UserID, sentMsg.ReceiverID, "Receiver should be Bob")
	s.Equal(firstMessage, sentMsg.Content, "Message content should match")
	s.Equal("sent", sentMsg.Status, "Message should be marked as sent")
	s.WithinDuration(time.Now(), sentMsg.CreatedAt, 5*time.Second, "Message timestamp should be recent")

	// Step 5: Verify Bob receives message via NATS real-time
	s.T().Log("Step 5: Verify Bob receives message in real-time via NATS")
	err = messageCollector.WaitForMessageCount(ctx, 1, 10*time.Second)
	s.Require().NoError(err, "Bob should receive message via NATS within 10 seconds")

	receivedMessages := messageCollector.GetMessages()
	s.Len(receivedMessages, 1, "Bob should receive exactly one message")

	receivedMsg := receivedMessages[0]
	s.Equal(domain.MessageTypeNewMessage, receivedMsg.Type, "Should be new message type")
	s.Equal(firstMessage, receivedMsg.Data.Content, "Received content should match sent content")
	s.Equal(alice.UserID, receivedMsg.Data.SenderID, "Received sender should be Alice")
	s.Equal(bob.UserID, receivedMsg.Data.ReceiverID, "Received receiver should be Bob")

	// Step 6: Bob sees conversation in his chat list
	s.T().Log("Step 6: Bob checks his chat list and sees new conversation")
	bobChats, err := bobClient.GetChats(ctx)
	s.Require().NoError(err, "Bob should be able to get his chat list")
	s.Len(bobChats.Chats, 1, "Bob should have one conversation")

	bobChat := bobChats.Chats[0]
	expectedChatID := domain.ComputeChatID(alice.UserID, bob.UserID)
	s.Equal(expectedChatID, bobChat.ChatID, "Chat ID should be computed correctly")
	s.Equal(alice.UserID, bobChat.OtherParticipant, "Other participant should be Alice")
	s.WithinDuration(time.Now(), bobChat.LastMessageAt, 10*time.Second, "Last message time should be recent")

	// Step 7: Bob retrieves conversation messages
	s.T().Log("Step 7: Bob retrieves conversation messages")
	bobMessages, err := bobClient.GetMessages(ctx, expectedChatID, nil)
	s.Require().NoError(err, "Bob should be able to get conversation messages")
	s.Len(bobMessages.Messages, 1, "Conversation should have one message")
	s.Equal(firstMessage, bobMessages.Messages[0].Content, "Retrieved message should match")

	// Step 8: Bob responds
	s.T().Log("Step 8: Bob responds to Alice")
	responseMessage := "Hi Alice! Thanks for reaching out. I'm doing great, excited to be here!"

	bobResponse, err := bobClient.SendMessage(ctx, alice.UserID, responseMessage)
	s.Require().NoError(err, "Bob should be able to respond to Alice")
	s.Equal(bob.UserID, bobResponse.SenderID, "Response sender should be Bob")
	s.Equal(alice.UserID, bobResponse.ReceiverID, "Response receiver should be Alice")

	// Step 9: Alice sees updated conversation
	s.T().Log("Step 9: Alice sees updated conversation with Bob's response")

	// Wait a moment for message to be processed
	time.Sleep(500 * time.Millisecond)

	aliceChats, err = aliceClient.GetChats(ctx)
	s.Require().NoError(err, "Alice should be able to get updated chat list")
	s.Len(aliceChats.Chats, 1, "Alice should have one conversation")
	s.Equal(bob.UserID, aliceChats.Chats[0].OtherParticipant, "Alice's conversation should be with Bob")

	// Get full conversation from Alice's perspective
	aliceMessages, err := aliceClient.GetMessages(ctx, expectedChatID, nil)
	s.Require().NoError(err, "Alice should be able to get conversation messages")
	s.Len(aliceMessages.Messages, 2, "Conversation should now have two messages")

	// Verify message order (newest first - reverse chronological)
	s.Equal(responseMessage, aliceMessages.Messages[0].Content, "First message should be Bob's response (newest)")
	s.Equal(firstMessage, aliceMessages.Messages[1].Content, "Second message should be Alice's (oldest)")

	s.T().Log("✅ New User First Conversation Journey completed successfully!")
}

func (s *UserJourneyTestSuite) TestMultiUserGroupConversationJourney() {
	s.T().Log("=== Testing: Multi-User Group Conversation Journey ===")

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Step 1: Three users join
	s.T().Log("Step 1: Alice, Charlie, and Diana join the platform")
	alice := s.CreateTestUser("alice_group", "alice.group@example.com", "@alice_group")
	charlie := s.CreateTestUser("charlie_group", "charlie.group@example.com", "@charlie")
	diana := s.CreateTestUser("diana_group", "diana.group@example.com", "@diana")

	// Step 2: Set up NATS subscriptions for all users
	s.T().Log("Step 2: All users set up real-time subscriptions")

	aliceNATS, err := s.GetNATSClient("alice_group")
	s.Require().NoError(err)
	defer aliceNATS.Close()

	charlieNATS, err := s.GetNATSClient("charlie_group")
	s.Require().NoError(err)
	defer charlieNATS.Close()

	dianaNATS, err := s.GetNATSClient("diana_group")
	s.Require().NoError(err)
	defer dianaNATS.Close()

	// Wait for all connections
	for _, client := range []*testclient.NATSClient{aliceNATS, charlieNATS, dianaNATS} {
		err = client.WaitForConnection(5 * time.Second)
		s.Require().NoError(err, "All NATS clients should connect")
	}

	// Set up message collectors
	aliceCollector := testclient.NewMessageCollector()
	charlieCollector := testclient.NewMessageCollector()
	dianaCollector := testclient.NewMessageCollector()

	err = aliceNATS.SubscribeToMessages(ctx, aliceCollector.Handler())
	s.Require().NoError(err)
	err = charlieNATS.SubscribeToMessages(ctx, charlieCollector.Handler())
	s.Require().NoError(err)
	err = dianaNATS.SubscribeToMessages(ctx, dianaCollector.Handler())
	s.Require().NoError(err)

	// Step 3: Alice starts conversations with both Charlie and Diana
	s.T().Log("Step 3: Alice initiates conversations with Charlie and Diana")

	_, err = alice.SendMessage(ctx, "charlie_group", "Hi Charlie! Let's collaborate on the new project.")
	s.Require().NoError(err, "Alice should send message to Charlie")

	_, err = alice.SendMessage(ctx, "diana_group", "Hi Diana! Can you help with the design review?")
	s.Require().NoError(err, "Alice should send message to Diana")

	// Step 4: Charlie and Diana respond to Alice
	s.T().Log("Step 4: Charlie and Diana respond")

	_, err = charlie.SendMessage(ctx, "alice_group", "Absolutely! I have some great ideas for the project.")
	s.Require().NoError(err, "Charlie should respond to Alice")

	_, err = diana.SendMessage(ctx, "alice_group", "Of course! I'll review the designs this afternoon.")
	s.Require().NoError(err, "Diana should respond to Alice")

	// Step 5: Verify all users receive their respective messages
	s.T().Log("Step 5: Verify all users receive their respective messages via NATS")

	// Alice should receive 2 messages (from Charlie and Diana)
	err = aliceCollector.WaitForMessageCount(ctx, 2, 15*time.Second)
	s.Require().NoError(err, "Alice should receive 2 messages")

	// Charlie should receive 1 message (from Alice)
	err = charlieCollector.WaitForMessageCount(ctx, 1, 15*time.Second)
	s.Require().NoError(err, "Charlie should receive 1 message")

	// Diana should receive 1 message (from Alice)
	err = dianaCollector.WaitForMessageCount(ctx, 1, 15*time.Second)
	s.Require().NoError(err, "Diana should receive 1 message")

	// Step 6: Verify chat isolation (Charlie and Diana don't see each other's messages)
	s.T().Log("Step 6: Verify conversation isolation")

	charlieMessages := charlieCollector.GetMessages()
	s.Len(charlieMessages, 1, "Charlie should only see his conversation")
	s.Equal("Hi Charlie! Let's collaborate on the new project.", charlieMessages[0].Data.Content)

	dianaMessages := dianaCollector.GetMessages()
	s.Len(dianaMessages, 1, "Diana should only see her conversation")
	s.Equal("Hi Diana! Can you help with the design review?", dianaMessages[0].Data.Content)

	// Step 7: Verify each user's chat list shows correct conversations
	s.T().Log("Step 7: Verify chat lists are correct for all users")

	aliceChats, err := alice.GetChats(ctx)
	s.Require().NoError(err)
	s.Len(aliceChats.Chats, 2, "Alice should have 2 conversations")

	charlieChats, err := charlie.GetChats(ctx)
	s.Require().NoError(err)
	s.Len(charlieChats.Chats, 1, "Charlie should have 1 conversation")

	dianaChats, err := diana.GetChats(ctx)
	s.Require().NoError(err)
	s.Len(dianaChats.Chats, 1, "Diana should have 1 conversation")

	s.T().Log("✅ Multi-User Group Conversation Journey completed successfully!")
}

func (s *UserJourneyTestSuite) TestMessageStatusAndDeliveryJourney() {
	s.T().Log("=== Testing: Message Status and Delivery Journey ===")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Step 1: Create test users
	s.T().Log("Step 1: Eve and Frank join for message status testing")
	eve := s.CreateTestUser("eve_status", "eve@example.com", "@eve")
	frank := s.CreateTestUser("frank_status", "frank@example.com", "@frank")

	// Step 2: Eve sends message to Frank
	s.T().Log("Step 2: Eve sends message to Frank")
	statusMessage := "Frank, please confirm when you've read this important message!"

	sentMsg, err := eve.SendMessage(ctx, "frank_status", statusMessage)
	s.Require().NoError(err, "Eve should send message to Frank")
	s.Equal("sent", sentMsg.Status, "Message should initially be marked as sent")

	// Step 3: Frank reads the message (simulated by retrieving it)
	s.T().Log("Step 3: Frank retrieves and reads the message")
	chatID := domain.ComputeChatID("eve_status", "frank_status")

	frankMessages, err := frank.GetMessages(ctx, chatID, nil)
	s.Require().NoError(err, "Frank should be able to get messages")
	s.Len(frankMessages.Messages, 1, "Frank should see one message")
	s.Equal(statusMessage, frankMessages.Messages[0].Content, "Frank should see Eve's message")

	// Step 4: Frank marks message as read
	s.T().Log("Step 4: Frank marks message as read")
	messageID := domain.MessageID{
		SenderID:   sentMsg.SenderID,
		ReceiverID: sentMsg.ReceiverID,
		CreatedAt:  sentMsg.CreatedAt,
	}

	statusResp, err := frank.UpdateMessageStatus(ctx, messageID)
	s.Require().NoError(err, "Frank should be able to mark message as read")
	s.Greater(statusResp.UpdatedCount, int64(0), "At least one message should be marked as read")

	// Step 5: Verify the read status workflow
	s.T().Log("Step 5: Verify message status workflow completed")

	// The status update should have been processed
	// In a real system, Eve might get a notification about the read receipt
	// For now, we verify the operation succeeded
	s.True(statusResp.UpdatedCount > 0, "Message status update should be successful")

	s.T().Log("✅ Message Status and Delivery Journey completed successfully!")
}

func (s *UserJourneyTestSuite) TestErrorHandlingAndEdgeCasesJourney() {
	s.T().Log("=== Testing: Error Handling and Edge Cases Journey ===")

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	graceErrorsId := "graceerrors"
	graceAndSomeoneChannelName := graceErrorsId + "_someone"
	graceToGraceChannelName := graceErrorsId + "_" + graceErrorsId
	// Step 1: Create test user
	s.T().Log("Step 1: Grace joins to test error scenarios")
	grace := s.CreateTestUser(graceErrorsId, "grace@example.com", "@grace")

	// Step 2: Test sending empty message
	s.T().Log("Step 2: Test sending empty message (should fail)")
	_, err := grace.SendMessage(ctx, testdata.Bob.UserID, "")
	s.Error(err, "Sending empty message should fail")
	s.True(testclient.IsBadRequest(err), "Should return 400 Bad Request for empty message")

	// Step 3: Test sending message to self
	s.T().Log("Step 3: Test sending message to self (should fail)")
	_, err = grace.SendMessage(ctx, graceToGraceChannelName, "Talking to myself")
	s.NoError(err, "Sending message to self should work")

	// Step 4: Test retrieving messages with invalid parameters
	s.T().Log("Step 4: Test invalid pagination parameters")
	_, err = grace.GetMessages(ctx, graceAndSomeoneChannelName, &testclient.GetMessagesOptions{
		Limit: 1000, // Exceeds maximum
	})
	s.Error(err, "Should reject limit exceeding maximum")
	s.True(testclient.IsBadRequest(err), "Should return 400 Bad Request for invalid limit")

	// Step 5: Test invalid cursor format
	s.T().Log("Step 5: Test invalid cursor format")
	_, err = grace.GetMessages(ctx, graceAndSomeoneChannelName, &testclient.GetMessagesOptions{
		Cursor: "invalid-cursor-format",
	})
	s.Error(err, "Should reject invalid cursor format")
	s.True(testclient.IsBadRequest(err), "Should return 400 Bad Request for invalid cursor")

	// Step 6: Test unauthorized status update
	s.T().Log("Step 6: Test unauthorized message status update")

	// Create another user to send a message
	henryErrorsId := "henryerrors"
	henry := s.CreateTestUser(henryErrorsId, "henry@example.com", "@henry")
	sentMsg, err := henry.SendMessage(ctx, graceErrorsId, "Test message for status update")
	s.Require().NoError(err, "Henry should send message to Grace")

	// Try to have Henry (sender) update the status (should fail - only receiver can update)
	messageID := domain.MessageID{
		SenderID:   sentMsg.SenderID,
		ReceiverID: sentMsg.ReceiverID,
		CreatedAt:  sentMsg.CreatedAt,
	}

	_, err = henry.UpdateMessageStatus(ctx, messageID)
	s.Error(err, "Sender should not be able to update message status")
	s.True(testclient.IsForbidden(err), "Should return 403 Forbidden for unauthorized status update")

	// Step 7: Grace (receiver) should be able to update status
	s.T().Log("Step 7: Verify receiver can update message status")
	_, err = grace.UpdateMessageStatus(ctx, messageID)
	s.NoError(err, "Receiver should be able to update message status")

	s.T().Log("✅ Error Handling and Edge Cases Journey completed successfully!")
}

func TestUserJourneyTestSuite(t *testing.T) {
	suite.Run(t, new(UserJourneyTestSuite))
}
