package domain

import (
	"regexp"
	"testing"
)

func TestNewProductID(t *testing.T) {
	result, err := NewProductID()
	expectation := "^PDT-[0-9A-Z]{6}$"
	if err != nil {
		t.Fatalf("newProductID: %v", err)
	}
	match := regexp.MustCompile(expectation).MatchString(result)
	if match == false {
		t.Errorf("Expected %s to equal %s\n", result, expectation)
	}
}

func TestNewCartID(t *testing.T) {
	result, err := NewCartID()
	expectation := "^BSK-[0-9A-Z]{6}$"
	if err != nil {
		t.Fatalf("newProductID: %v", err)
	}
	match := regexp.MustCompile(expectation).MatchString(result)
	if match == false {
		t.Errorf("Expected %s to equal %s\n", result, expectation)
	}
}

func TestNewOrderID(t *testing.T) {
	result, err := NewOrderID()
	expectation := "^CMD-[0-9A-Z]{6}$"
	if err != nil {
		t.Fatalf("newProductID: %v", err)
	}
	match := regexp.MustCompile(expectation).MatchString(result)
	if match == false {
		t.Errorf("Expected %s to equal %s\n", result, expectation)
	}
}

func TestGenerateSuffixUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id, err := generateSuffix(6)
		if err != nil {
			t.Fatalf("NewProductID: %v", err)
		}
		if seen[id] {
			t.Errorf("Duplicate ID detected: %s", id)
		}
		seen[id] = true
	}
}