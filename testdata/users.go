package testdata

import "messaging-app/internal/domain"

// User variables for consistent referencing
var (
	Alice = domain.UserContext{
		UserID:  "alice",
		Email:   "alice@interface.ai",
		Handler: "alice_dev",
	}

	Bob = domain.UserContext{
		UserID:  "bob",
		Email:   "bob@interface.ai",
		Handler: "bob_product",
	}

	Charlie = domain.UserContext{
		UserID:  "charlie",
		Email:   "charlie@interface.ai",
		Handler: "charlie_design",
	}

	Diana = domain.UserContext{
		UserID:  "diana",
		Email:   "diana@interface.ai",
		Handler: "diana_eng",
	}

	Eve = domain.UserContext{
		UserID:  "eve",
		Email:   "eve@interface.ai",
		Handler: "eve_qa",
	}

	// Invalid users for validation testing
	EmptyUserID = domain.UserContext{
		UserID:  "",
		Email:   "empty@example.com",
		Handler: "empty_user",
	}

	EmptyEmail = domain.UserContext{
		UserID:  "valid_user",
		Email:   "",
		Handler: "no_email",
	}

	EmptyHandler = domain.UserContext{
		UserID:  "valid_user",
		Email:   "valid@example.com",
		Handler: "",
	}

	WhitespaceUserID = domain.UserContext{
		UserID:  "   ",
		Email:   "whitespace@example.com",
		Handler: "whitespace_user",
	}

	UnicodeUser = domain.UserContext{
		UserID:  "user_with_unicode_🚀",
		Email:   "unicode@example.com",
		Handler: "unicode_user",
	}

	EmojiUser = domain.UserContext{
		UserID:  "user_with_😀_emoji",
		Email:   "emoji@example.com",
		Handler: "emoji_user_🎉",
	}
)

// Edge case users (valid but unusual)
var (
	DashUser = domain.UserContext{
		UserID:  "user-with-dashes",
		Email:   "dashes@example.com",
		Handler: "dash-user",
	}

	DotUser = domain.UserContext{
		UserID:  "user.with.dots",
		Email:   "dots@example.com",
		Handler: "dot.user",
	}

	NumberUser = domain.UserContext{
		UserID:  "user_with_numbers123",
		Email:   "numbers123@example.com",
		Handler: "num123",
	}

	SingleCharUser = domain.UserContext{
		UserID:  "a",
		Email:   "single@example.com",
		Handler: "a",
	}

	UppercaseUser = domain.UserContext{
		UserID:  "UPPERCASE_USER",
		Email:   "UPPER@EXAMPLE.COM",
		Handler: "UPPER",
	}
)

// ValidUsers returns a collection of valid user contexts for testing
func ValidUsers() []domain.UserContext {
	return []domain.UserContext{
		Alice,
		Bob,
		Charlie,
		Diana,
		Eve,
	}
}

// InvalidUsers returns users that should fail validation
func InvalidUsers() []domain.UserContext {
	return []domain.UserContext{
		EmptyUserID,
		EmptyEmail,
		EmptyHandler,
		WhitespaceUserID,
		UnicodeUser,
		EmojiUser,
	}
}

// EdgeCaseUsers returns users with edge case characteristics (valid but unusual)
func EdgeCaseUsers() []domain.UserContext {
	return []domain.UserContext{
		DashUser,
		DotUser,
		NumberUser,
		SingleCharUser,
		UppercaseUser,
	}
}

// GetUserByID returns a specific user from valid users
func GetUserByID(userID string) *domain.UserContext {
	for _, user := range ValidUsers() {
		if user.UserID == userID {
			return &user
		}
	}
	return nil
}