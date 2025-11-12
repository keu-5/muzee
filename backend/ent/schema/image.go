package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Image holds the schema definition for the Image entity.
type Image struct {
	ent.Schema
}

// Fields of the Image.
func (Image) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Unique().
			Immutable(),

		field.Int64("post_id"),

		field.String("image_url").
			MaxLen(255).
			Immutable(),

		field.Int8("display_order").
			NonNegative(),

		field.Int16("width").
			Immutable().
			NonNegative(),

		field.Int16("height").
			Immutable().
			NonNegative(),

		field.Text("alt_text").
			Optional(),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Image.
func (Image) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("post", Post.Type).
			Ref("images").
			Field("post_id").
			Unique().
			Required(),
	}
}

func (Image) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("post_id"),
		index.Fields("post_id", "display_order").
			Unique().
			StorageKey("idx_images_post_display_order"),
	}
}
