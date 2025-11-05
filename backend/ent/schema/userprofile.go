package schema

import "entgo.io/ent"

// UserProfile holds the schema definition for the UserProfile entity.
type UserProfile struct {
	ent.Schema
}

// Fields of the UserProfile.
func (UserProfile) Fields() []ent.Field {
	return nil
}

// Edges of the UserProfile.
func (UserProfile) Edges() []ent.Edge {
	return nil
}
