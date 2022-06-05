package main

import (
	"reflect"
	"testing"

	"github.com/aquilax/listboard/node"
)

func TestSession_GetNodeURL(t *testing.T) {
	tests := []struct {
		name string
		n    *node.Node
		want string
	}{
		{
			"node level 1",
			&node.Node{
				ID:       "1",
				ParentID: node.RootNodeID,
				Title:    "Test Node",
				DomainID: "1",
				Level:    levelRoot,
				TripCode: getTripCode("test"),
			},
			"http://www.example.com/list/1/test-node.html",
		},
		{
			"node level 2",
			&node.Node{
				ID:       "2",
				ParentID: "1",
				Title:    "Test Node",
				DomainID: "1",
				Level:    levelList,
				TripCode: getTripCode("test"),
			},
			"http://www.example.com/vote/2/test-node.html#post",
		},
		{
			"node level 3",
			&node.Node{
				ID:       "3",
				ParentID: "2",
				Title:    "Test Node",
				DomainID: "1",
				Level:    levelVote,
				TripCode: getTripCode("test"),
			},
			"http://www.example.com/vote/2/item.html#I3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := getTestConfig()
			sc := c.getSiteConfig("")
			s := NewSession(sc, nil)
			if got := s.GetNodeURL(tt.n).String(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Session.GetNodeURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
