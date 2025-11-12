package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Like holds the schema definition for the Like entity.
type Like struct {
	ent.Schema
}

// Fields of the Like.
func (Like) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Unique().
			Immutable(),

		field.Int64("user_id"),

		field.Int64("post_id"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Like.
func (Like) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("likes").
			Field("user_id").
			Unique().
			Required().
			StructTag(`json:"user,omitempty"`),

		edge.From("post", Post.Type).
			Ref("likes").
			Field("post_id").
			Unique().
			Required().
			StructTag(`json:"post,omitempty"`),
	}
}

func (Like) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "post_id").
			Unique().
			StorageKey("idx_likes_user_post"),
		index.Fields("post_id"),
		index.Fields("user_id", "created_at").
			StorageKey("idx_likes_user_created"),
	}
}