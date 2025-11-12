package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Post holds the schema definition for the Post entity.
type Post struct {
	ent.Schema
}

// Fields of the Post.
func (Post) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Unique().
			Immutable(),

		field.Int64("user_id"),

		field.Int64("recommended_post_id").
			Optional(),

		field.Text("caption").
			Optional(),

		field.Bool("is_public").
			Default(true),

		field.Bool("is_nsfw").
			Default(false),

		field.Int("liked_count").
			Default(0).
			NonNegative(),

		field.Int("recommended_count").
			Default(0).
			NonNegative(),

		field.Int("view_count").
			Default(0).
			NonNegative(),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Post.
func (Post) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("posts").
			Field("user_id").
			Unique().
			Required(),
		
		edge.To("recommended_post", Post.Type).
			Field("recommended_post_id").
			Unique().
			From("recommending_posts").
			StructTag(`json:"recommended_post,omitempty"`),

		edge.To("likes", Like.Type).
			StructTag(`json:"likes,omitempty"`),
	}
}

func (Post) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("created_at").
			StorageKey("idx_posts_created_at"),
		index.Fields("recommended_post_id"),
		index.Fields("user_id", "created_at").
			StorageKey("idx_posts_user_created"),
	}
}