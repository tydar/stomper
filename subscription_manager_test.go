package main

import (
	"fmt"
	"testing"
	// "reflect"

	"github.com/google/uuid"
)

func TestSubscriptionManager(t *testing.T) {
	sm := NewSubscriptionManager()
	clientID := uuid.NewString()
	subID := "test"
	dest := "/queue/test"

	// sm.Subscribe
	t.Run("_Subscribe", func(t *testing.T) {
		err := sm.Subscribe(clientID, subID, dest)
		if err != nil {
			t.Errorf("Subscription error: %v\n", err)
		}
		v, prs := sm.Subscriptions[clientID+"_"+subID]
		if !prs {
			t.Errorf("No subscription added.")
		} else if v.Destination != dest {
			t.Errorf("Subscription added incorrectly.")
		}

		err = sm.Subscribe(clientID, subID, dest)
		if err == nil {
			t.Errorf("Duplicate subscription allowed.")
		}
	})

	// sm.Unsubscribe
	t.Run("_Unsubscribe", func(t *testing.T) {
		err := sm.Unsubscribe(clientID, subID)
		if err != nil {
			t.Errorf("Unsubscribe error: %v\n", err)
		}

		err = sm.Unsubscribe(clientID, subID)
		if err == nil {
			t.Errorf("Duplicate unsubscribe allowed.")
		}
	})

	// sm.Get
	t.Run("_Get", func(t *testing.T) {
		err := sm.Subscribe(clientID, subID, dest)
		if err != nil {
			t.Errorf("Subscription error: %v\n", err)
		}

		s, err := sm.Get(clientID, subID)
		if err != nil {
			t.Errorf("Get error: %v\n", err)
		}

		if s.ClientID != clientID {
			t.Errorf("Retrieved incorrect sub, somehow: got %s wanted %s\n", s.ClientID, clientID)
		}
	})
}

func TestSubscriptionManagerGetByDestination(t *testing.T) {
	sm := NewSubscriptionManager()
	sub1to1 := Subscription{ID: "test1", Destination: "/queue/test1", ClientID: "test1"}
	sub1to2 := Subscription{ID: "test2", Destination: "/queue/test2", ClientID: "test1"}
	sub2to1 := Subscription{ID: "test1", Destination: "/queue/test1", ClientID: "test2"}
	testDest := "/queue/test1"
	var tests = []struct {
		subs []Subscription
		want int
	}{
		{subs: []Subscription{}, want: 0},
		{subs: []Subscription{sub1to1}, want: 1},
		{subs: []Subscription{sub1to1, sub1to2}, want: 1},
		{subs: []Subscription{sub1to1, sub2to1}, want: 2},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%v\n", tt.subs)
		t.Run(testname, func(t *testing.T) {
			for _, sub := range tt.subs {
				err := sm.Subscribe(sub.ClientID, sub.ID, sub.Destination)
				if err != nil {
					t.Errorf("Subscription error: %v\n", err)
				}
				defer sm.Unsubscribe(sub.ClientID, sub.ID) // defer for test cleanup between rounds
			}
			result := sm.ClientsByDestination(testDest)
			if len(result) != tt.want {
				t.Errorf("Wrong client count: got %d wanted %d\n", len(result), tt.want)
			}
		})
	}
}
