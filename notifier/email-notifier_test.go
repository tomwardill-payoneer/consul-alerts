package notifier

import (
	"fmt"
	"net/smtp"
	"reflect"
	"strings"
	"testing"
)

func TestNotify(t *testing.T) {
	oldSendMail := sendMail
	defer func() {
		sendMail = oldSendMail
	}()

	host := "mailserver.localdomain"
	port := 123

	expectedAddr := fmt.Sprintf("%s:%d", host, port)
	expectedFrom := "sender@example.com"
	expectedTo := []string{"test1@example.com", "test2@example.com"}
	expectedMsg := `From: "Some Sender" <sender@example.com>
To: test1@example.com, test2@example.com
Subject: Some Cluster is HEALTHY
MIME-version: 1.0;
Content-Type: text/html; charset="UTF-8";


<!DOCTYPE html>
`

	sendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		if addr != expectedAddr {
			t.Errorf("expected %s, got %s", expectedAddr, addr)
		}

		if a == nil {
			t.Error("auth must not be null")
		}

		if from != expectedFrom {
			t.Errorf("expected %s, got %s", expectedFrom, from)
		}

		if !reflect.DeepEqual(to, expectedTo) {
			t.Errorf("expected %s, got %s", expectedTo, to)
		}

		stringMsg := string(msg)
		if !strings.HasPrefix(stringMsg, expectedMsg) {
			t.Errorf("expected message to start with\n\n%s\n\ngot\n\n%s", expectedMsg, stringMsg)
		}

		return nil
	}

	notifier := EmailNotifier{
		Username:    "some username",
		Password:    "some password",
		ClusterName: "Some Cluster",
		Url:         host,
		Port:        port,
		SenderEmail: expectedFrom,
		SenderAlias: "Some Sender",
		Receivers:   expectedTo,
	}

	if !notifier.Notify(make(Messages, 0)) {
		t.Error("Notify must return true")
	}
}
func TestNotifyWithoutAuth(t *testing.T) {
	oldSendMail := sendMail
	defer func() {
		sendMail = oldSendMail
	}()

	host := "mailserver.localdomain"
	port := 123

	expectedAddr := fmt.Sprintf("%s:%d", host, port)
	expectedFrom := "sender@example.com"
	expectedTo := []string{"test1@example.com"}

	sendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		if addr != expectedAddr {
			t.Errorf("expected %s, got %s", expectedAddr, addr)
		}

		if a != nil {
			t.Error("auth must be null when no username/password provided")
		}

		if from != expectedFrom {
			t.Errorf("expected %s, got %s", expectedFrom, from)
		}

		if !reflect.DeepEqual(to, expectedTo) {
			t.Errorf("expected %s, got %s", expectedTo, to)
		}

		return nil
	}

	notifier := EmailNotifier{
		ClusterName: "Some Cluster",
		Url:         host,
		Port:        port,
		SenderEmail: expectedFrom,
		SenderAlias: "Some Sender",
		Receivers:   expectedTo,
	}

	if !notifier.Notify(make(Messages, 0)) {
		t.Error("Notify must return true")
	}
}

func TestNotifyOnePerAlert(t *testing.T) {
	oldSendMail := sendMail
	defer func() {
		sendMail = oldSendMail
	}()

	callCount := 0
	expectedCallCount := 2

	sendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		callCount++
		stringMsg := string(msg)

		if callCount == 1 && !strings.Contains(stringMsg, "node1 - check1") {
			t.Error("First email should contain node1 - check1")
		}
		if callCount == 2 && !strings.Contains(stringMsg, "node2 - check2") {
			t.Error("Second email should contain node2 - check2")
		}

		return nil
	}

	notifier := EmailNotifier{
		Username:    "user",
		Password:    "pass",
		ClusterName: "Cluster",
		Url:         "localhost",
		Port:        25,
		SenderEmail: "sender@example.com",
		SenderAlias: "Sender",
		Receivers:   []string{"receiver@example.com"},
		OnePerAlert: true,
	}

	alerts := Messages{
		{Node: "node1", CheckId: "check1", Status: "passing"},
		{Node: "node2", CheckId: "check2", Status: "passing"},
	}

	if !notifier.Notify(alerts) {
		t.Error("Notify must return true")
	}

	if callCount != expectedCallCount {
		t.Errorf("expected %d emails, got %d", expectedCallCount, callCount)
	}
}

func TestNotifyOnePerNode(t *testing.T) {
	oldSendMail := sendMail
	defer func() {
		sendMail = oldSendMail
	}()

	callCount := 0
	expectedCallCount := 2

	sendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		callCount++
		stringMsg := string(msg)

		if callCount == 1 && !strings.Contains(stringMsg, "Cluster node1") {
			t.Error("First email should contain node1")
		}
		if callCount == 2 && !strings.Contains(stringMsg, "Cluster node2") {
			t.Error("Second email should contain node2")
		}

		return nil
	}

	notifier := EmailNotifier{
		Username:    "user",
		Password:    "pass",
		ClusterName: "Cluster",
		Url:         "localhost",
		Port:        25,
		SenderEmail: "sender@example.com",
		SenderAlias: "Sender",
		Receivers:   []string{"receiver@example.com"},
		OnePerNode:  true,
	}

	alerts := Messages{
		{Node: "node1", CheckId: "check1", Status: "passing"},
		{Node: "node1", CheckId: "check2", Status: "passing"},
		{Node: "node2", CheckId: "check3", Status: "passing"},
	}

	if !notifier.Notify(alerts) {
		t.Error("Notify must return true")
	}

	if callCount != expectedCallCount {
		t.Errorf("expected %d emails, got %d", expectedCallCount, callCount)
	}
}

func TestNotifyReturnsFailureOnError(t *testing.T) {
	oldSendMail := sendMail
	defer func() {
		sendMail = oldSendMail
	}()

	sendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		return fmt.Errorf("simulated error")
	}

	notifier := EmailNotifier{
		Username:    "user",
		Password:    "pass",
		ClusterName: "Cluster",
		Url:         "localhost",
		Port:        25,
		SenderEmail: "sender@example.com",
		SenderAlias: "Sender",
		Receivers:   []string{"receiver@example.com"},
	}

	if notifier.Notify(make(Messages, 0)) {
		t.Error("Notify must return false on error")
	}
}

func TestNotifierName(t *testing.T) {
	notifier := EmailNotifier{}
	if notifier.NotifierName() != "email" {
		t.Errorf("expected 'email', got '%s'", notifier.NotifierName())
	}
}

func TestCopy(t *testing.T) {
	original := EmailNotifier{
		ClusterName: "Test Cluster",
		Enabled:     true,
		Username:    "user",
		Password:    "pass",
	}

	copy := original.Copy()

	if !reflect.DeepEqual(&original, copy) {
		t.Error("Copy should create an identical notifier")
	}

	// Verify it's a different instance
	copyNotifier := copy.(*EmailNotifier)
	copyNotifier.ClusterName = "Modified"
	if original.ClusterName == "Modified" {
		t.Error("Modifying copy should not affect original")
	}
}
