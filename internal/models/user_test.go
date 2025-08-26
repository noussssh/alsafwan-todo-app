package models

import (
	"testing"
	"time"
)

func TestUserSetPassword(t *testing.T) {
	user := &User{}
	password := "testpassword123"
	
	err := user.SetPassword(password)
	if err != nil {
		t.Fatalf("SetPassword failed: %v", err)
	}
	
	if user.PasswordDigest == "" {
		t.Error("Password digest should not be empty")
	}
	
	if user.PasswordResetAt == nil {
		t.Error("PasswordResetAt should be set")
	}
	
	if user.PasswordExpiresAt == nil {
		t.Error("PasswordExpiresAt should be set")
	}
}

func TestUserCheckPassword(t *testing.T) {
	user := &User{}
	password := "testpassword123"
	
	user.SetPassword(password)
	
	if !user.CheckPassword(password) {
		t.Error("CheckPassword should return true for correct password")
	}
	
	if user.CheckPassword("wrongpassword") {
		t.Error("CheckPassword should return false for incorrect password")
	}
}

func TestUserIsPasswordExpired(t *testing.T) {
	user := &User{}
	
	if user.IsPasswordExpired() {
		t.Error("Password should not be expired when PasswordExpiresAt is nil")
	}
	
	pastTime := time.Now().Add(-1 * time.Hour)
	user.PasswordExpiresAt = &pastTime
	
	if !user.IsPasswordExpired() {
		t.Error("Password should be expired when PasswordExpiresAt is in the past")
	}
	
	futureTime := time.Now().Add(1 * time.Hour)
	user.PasswordExpiresAt = &futureTime
	
	if user.IsPasswordExpired() {
		t.Error("Password should not be expired when PasswordExpiresAt is in the future")
	}
}

func TestUserUpdateSignInInfo(t *testing.T) {
	user := &User{SignInCount: 5}
	initialTime := time.Now().Add(-1 * time.Hour)
	user.CurrentSignInAt = &initialTime
	
	user.UpdateSignInInfo()
	
	if user.SignInCount != 6 {
		t.Errorf("SignInCount should be 6, got %d", user.SignInCount)
	}
	
	if user.LastSignInAt == nil {
		t.Error("LastSignInAt should not be nil")
	}
	
	if user.CurrentSignInAt == nil {
		t.Error("CurrentSignInAt should not be nil")
	}
	
	if !user.LastSignInAt.Equal(initialTime) {
		t.Error("LastSignInAt should equal the previous CurrentSignInAt")
	}
}

func TestUserCanManageUser(t *testing.T) {
	admin := &User{Role: RoleAdmin}
	manager := &User{Role: RoleManager}
	salesperson := &User{Role: RoleSalesperson}
	
	if !admin.CanManageUser(manager) {
		t.Error("Admin should be able to manage manager")
	}
	
	if !admin.CanManageUser(salesperson) {
		t.Error("Admin should be able to manage salesperson")
	}
	
	if manager.CanManageUser(admin) {
		t.Error("Manager should not be able to manage admin")
	}
	
	if !manager.CanManageUser(salesperson) {
		t.Error("Manager should be able to manage salesperson")
	}
	
	if salesperson.CanManageUser(admin) {
		t.Error("Salesperson should not be able to manage admin")
	}
}

func TestUserCanDisableUser(t *testing.T) {
	admin := &User{ID: 1, Role: RoleAdmin}
	manager := &User{ID: 2, Role: RoleManager}
	salesperson := &User{ID: 3, Role: RoleSalesperson}
	
	if admin.CanDisableUser(admin) {
		t.Error("User should not be able to disable themselves")
	}
	
	if admin.CanDisableUser(manager) {
		t.Error("Should not be able to disable non-salesperson")
	}
	
	if !admin.CanDisableUser(salesperson) {
		t.Error("Admin should be able to disable salesperson")
	}
}

func TestGenerateStrongPassword(t *testing.T) {
	password, err := GenerateStrongPassword()
	if err != nil {
		t.Fatalf("GenerateStrongPassword failed: %v", err)
	}
	
	if len(password) != 8 {
		t.Errorf("Password should be 8 characters long, got %d", len(password))
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		password string
		valid    bool
	}{
		{"", false},
		{"12345", false},
		{"123456", true},
		{"longenoughpassword", true},
	}
	
	for _, test := range tests {
		err := ValidatePassword(test.password)
		if test.valid && err != nil {
			t.Errorf("Password %q should be valid", test.password)
		}
		if !test.valid && err == nil {
			t.Errorf("Password %q should be invalid", test.password)
		}
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"", false},
		{"J", false},
		{"Jo", true},
		{"John Doe", true},
		{longString(101), false},
	}
	
	for _, test := range tests {
		err := ValidateName(test.name)
		if test.valid && err != nil {
			t.Errorf("Name %q should be valid", test.name)
		}
		if !test.valid && err == nil {
			t.Errorf("Name %q should be invalid", test.name)
		}
	}
}

func TestValidateCompany(t *testing.T) {
	tests := []struct {
		company *string
		valid   bool
	}{
		{nil, true},
		{stringPtr("Al Safwan Marine"), true},
		{stringPtr("Louis Safety"), true},
		{stringPtr("Data Grid Labs"), true},
		{stringPtr("Invalid Company"), false},
	}
	
	for _, test := range tests {
		err := ValidateCompany(test.company)
		if test.valid && err != nil {
			t.Errorf("Company %v should be valid", test.company)
		}
		if !test.valid && err == nil {
			t.Errorf("Company %v should be invalid", test.company)
		}
	}
}

func longString(n int) string {
	result := make([]byte, n)
	for i := range result {
		result[i] = 'a'
	}
	return string(result)
}

func stringPtr(s string) *string {
	return &s
}