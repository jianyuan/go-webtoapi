package main

import (
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

func newSelectorField(fn func(*goquery.Selection) (interface{}, error)) *graphql.Field {
	return &graphql.Field{
		Type: graphql.String,
		Args: graphql.FieldConfigArgument{
			"selector": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			selection, ok := p.Source.(*goquery.Selection)
			if ok {
				selector, ok := p.Args["selector"].(string)
				if ok {
					selection = selection.Find(selector)
				}
				return fn(selection)
			}
			return nil, nil
		},
	}
}

var DocumentType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Document",
	Fields: graphql.Fields{
		"content": newSelectorField(func(selection *goquery.Selection) (interface{}, error) {
			return selection.Html()
		}),
		"html": newSelectorField(func(selection *goquery.Selection) (interface{}, error) {
			return goquery.OuterHtml(selection)
		}),
		"text": newSelectorField(func(selection *goquery.Selection) (interface{}, error) {
			return selection.Text(), nil
		}),
		"tag": newSelectorField(func(selection *goquery.Selection) (interface{}, error) {
			return goquery.NodeName(selection), nil
		}),
	},
})

func main() {
	fields := graphql.Fields{
		"page": &graphql.Field{
			Type: DocumentType,
			Args: graphql.FieldConfigArgument{
				"url": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				url, ok := p.Args["url"].(string)
				if ok {
					resp, err := http.Get(url)
					if err != nil {
						return nil, err
					}
					defer resp.Body.Close()

					doc, err := goquery.NewDocumentFromReader(resp.Body)
					if err != nil {
						return nil, err
					}

					return doc.Selection, nil
				}
				return nil, nil
			},
		},
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, _ := graphql.NewSchema(schemaConfig)

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	http.Handle("/graphql", h)
	http.ListenAndServe(":8080", nil)
}
