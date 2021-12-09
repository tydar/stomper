package main

import (
	"fmt"
	"log"
)

type SubscriptionManager struct {
	Subscriptions map[string]Subscription
}

func NewSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		Subscriptions: make(map[string]Subscription),
	}
}

func (sm *SubscriptionManager) Subscribe(clientID string, subID string, dest string) error {
	internalSubID := clientID + "_" + subID
	_, prs := sm.Subscriptions[internalSubID]
	if prs {
		return fmt.Errorf("subscription from client %s with sub ID %s already exists", clientID, subID)
	}

	sm.Subscriptions[internalSubID] = Subscription{
		ID:          subID,
		Destination: dest,
		ClientID:    clientID,
	}
	log.Printf("NEW_SUBSCRIPTION: Sub %s from client %s to dest %s\n", subID, clientID, dest)
	return nil
}

func (sm *SubscriptionManager) Unsubscribe(clientID string, subID string) error {
	internalSubID := clientID + "_" + subID
	sub, prs := sm.Subscriptions[internalSubID]
	if !prs {
		return fmt.Errorf("no such subscription %s for client %s", subID, clientID)
	}
	log.Printf("UNSUBSCRIBE: sub %s from client %s to dest %s\n", subID, clientID, sub.Destination)
	delete(sm.Subscriptions, internalSubID)
	return nil
}

func (sm *SubscriptionManager) Get(clientID string, subID string) (Subscription, error) {
	internalSubID := clientID + "_" + subID
	sub, prs := sm.Subscriptions[internalSubID]
	if prs {
		return sub, nil
	} else {
		return Subscription{}, fmt.Errorf("no sub ID %s for client %s exists", subID, clientID)
	}
}

func (sm *SubscriptionManager) ClientsByDestination(dest string) []Subscription {
	// loop for now
	// could see this being expensive
	// might be better to just maintain a map keyed by internal ID
	// and one keyed by destination
	// so lookup is O(1) or O(nlogn) or something like that
	// adding is worst-case O(n), and removing is O(1)

	clients := make([]Subscription, 0)
	for k := range sm.Subscriptions {
		sub := sm.Subscriptions[k]
		if sub.Destination == dest {
			clients = append(clients, sub)
		}
	}
	return clients
}

type Subscription struct {
	ID          string
	Destination string
	ClientID    string
}

func (s *Subscription) InternalSubID() string {
	return s.ClientID + "_" + s.ID
}
