package main

import (
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

func selectionFromResolveParams(p graphql.ResolveParams) (*goquery.Selection, bool) {
	selection, ok := p.Source.(*goquery.Selection)
	if ok {
		selector, ok := p.Args["selector"].(string)
		if ok {
			selection = selection.Find(selector)
		}
		return selection, true
	}
	return nil, false
}

func newSelectorField(fn func(*goquery.Selection) (interface{}, error)) *graphql.Field {
	return &graphql.Field{
		Type: graphql.String,
		Args: graphql.FieldConfigArgument{
			"selector": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			selection, ok := selectionFromResolveParams(p)
			if ok {
				return fn(selection)
			}
			return nil, nil
		},
	}
}

var NodeType *graphql.Interface
var ElementType *graphql.Object
var DocumentType *graphql.Object

func addNodeFieldConfigs(gt *graphql.Object) {
	gt.AddFieldConfig("content", newSelectorField(func(selection *goquery.Selection) (interface{}, error) {
		return selection.First().Html()
	}))
	gt.AddFieldConfig("html", newSelectorField(func(selection *goquery.Selection) (interface{}, error) {
		return goquery.OuterHtml(selection)
	}))
	gt.AddFieldConfig("text", newSelectorField(func(selection *goquery.Selection) (interface{}, error) {
		return selection.First().Text(), nil
	}))
	gt.AddFieldConfig("tag", newSelectorField(func(selection *goquery.Selection) (interface{}, error) {
		return goquery.NodeName(selection), nil
	}))
	gt.AddFieldConfig("attr", &graphql.Field{
		Type: graphql.String,
		Args: graphql.FieldConfigArgument{
			"selector": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
			"name": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			selection, ok := selectionFromResolveParams(p)
			if ok {
				name, ok := p.Args["name"].(string)
				if ok {
					val, exists := selection.Attr(name)
					if exists {
						return val, nil
					}
				}
			}
			return nil, nil
		},
	})
	gt.AddFieldConfig("query", &graphql.Field{
		Type: graphql.NewList(ElementType),
		Args: graphql.FieldConfigArgument{
			"selector": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.String),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			selection, ok := selectionFromResolveParams(p)
			if ok {
				var items []interface{}
				selection.Each(func(_ int, item *goquery.Selection) {
					items = append(items, item)
				})
				return items, nil
			}
			return nil, nil
		},
	})
	gt.AddFieldConfig("next", &graphql.Field{
		Type: ElementType,
		Args: graphql.FieldConfigArgument{
			"selector": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			selection, ok := selectionFromResolveParams(p)
			if ok {
				return selection.Next(), nil
			}
			return nil, nil
		},
	})
	gt.AddFieldConfig("nextAll", &graphql.Field{
		Type: graphql.NewList(ElementType),
		Args: graphql.FieldConfigArgument{
			"selector": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			selection, ok := selectionFromResolveParams(p)
			if ok {
				var items []interface{}
				selection.NextAll().Each(func(_ int, item *goquery.Selection) {
					items = append(items, item)
				})
				return items, nil
			}
			return nil, nil
		},
	})
}

func init() {
	NodeType = graphql.NewInterface(graphql.InterfaceConfig{
		Name: "Node",
		Fields: graphql.FieldsThunk(func() graphql.Fields {
			return graphql.Fields{
				"content": &graphql.Field{
					Type: graphql.String,
					Args: graphql.FieldConfigArgument{
						"selector": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
					},
				},
				"html": &graphql.Field{
					Type: graphql.String,
					Args: graphql.FieldConfigArgument{
						"selector": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
					},
				},
				"text": &graphql.Field{
					Type: graphql.String,
					Args: graphql.FieldConfigArgument{
						"selector": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
					},
				},
				"tag": &graphql.Field{
					Type: graphql.String,
					Args: graphql.FieldConfigArgument{
						"selector": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
					},
				},
				"attr": &graphql.Field{
					Type: graphql.String,
					Args: graphql.FieldConfigArgument{
						"selector": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
						"name": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
					},
				},
				"query": &graphql.Field{
					Type: graphql.NewList(ElementType),
					Args: graphql.FieldConfigArgument{
						"selector": &graphql.ArgumentConfig{
							Type: graphql.NewNonNull(graphql.String),
						},
					},
				},
				"next": &graphql.Field{
					Type: ElementType,
					Args: graphql.FieldConfigArgument{
						"selector": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
					},
				},
				"nextAll": &graphql.Field{
					Type: graphql.NewList(ElementType),
					Args: graphql.FieldConfigArgument{
						"selector": &graphql.ArgumentConfig{
							Type: graphql.String,
						},
					},
				},
			}
		}),
	})

	DocumentType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Document",
		Interfaces: []*graphql.Interface{
			NodeType,
		},
		Fields: graphql.Fields{
			"title": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					selection, ok := p.Source.(*goquery.Selection)
					if ok {
						return selection.Find("title").Text(), nil
					}
					return nil, nil
				},
			},
		},
	})

	ElementType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Element",
		Interfaces: []*graphql.Interface{
			NodeType,
		},
		Fields: graphql.Fields{},
	})

	addNodeFieldConfigs(DocumentType)
	addNodeFieldConfigs(ElementType)
}

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
	schema, err := graphql.NewSchema(schemaConfig)

	if err != nil {
		log.Fatal(err)
	}

	h := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	http.Handle("/graphql", h)
	http.ListenAndServe(":8080", nil)
}
