package main

import (
	"github.com/supabase-community/supabase-go"
)

func NewSupabaseClient(SUPABASE_API_URL, SUPABASE_API_KEY string) *supabase.Client {
	client, err := supabase.NewClient(SUPABASE_API_URL, SUPABASE_API_KEY, &supabase.ClientOptions{})
	if err != nil {
		panic(err)
	}
	return client
}
