package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgresql://postgres.ujfqzyixckoqbbvsakbg:2d6qF69QpeGzKnU6@aws-0-eu-west-1.pooler.supabase.com:5432/postgres?sslmode=require")
	if err != nil {
		log.Fatal(err)
	}

	// Check super_admins table
	fmt.Println("=== super_admins table ===")
	rows, err := db.QueryContext(context.Background(), "SELECT * FROM super_admins LIMIT 10")
	if err != nil {
		fmt.Printf("super_admins table error: %v\n", err)
	} else {
		cols, _ := rows.Columns()
		fmt.Printf("Columns: %v\n", cols)
		for rows.Next() {
			vals := make([]interface{}, len(cols))
			ptrs := make([]interface{}, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			rows.Scan(ptrs...)
			fmt.Printf("Row: %v\n", vals)
		}
		rows.Close()
	}

	// Check company details
	fmt.Println("\n=== company details ===")
	var adminEmail, subStatus, planType, authStatus string
	var subExpiry sql.NullTime
	err = db.QueryRowContext(context.Background(),
		"SELECT admin_email, subscription_status, subscription_expiry, plan_type, auth_status FROM companies WHERE admin_email = $1",
		"emmanuelztrd@gmail.com").Scan(&adminEmail, &subStatus, &subExpiry, &planType, &authStatus)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Email: %s\nSub Status: %s\nSub Expiry: %v\nPlan Type: %s\nAuth Status: %s\n", adminEmail, subStatus, subExpiry, planType, authStatus)

	// Check GetAllActiveCompanies query result
	fmt.Println("\n=== GetAllActiveCompanies result ===")
	rows2, err := db.QueryContext(context.Background(),
		"SELECT id, name, admin_email, auth_status, subscription_status, subscription_expiry FROM companies WHERE auth_status = 'active' AND subscription_status IN ('active', 'trialing') AND (subscription_expiry IS NULL OR subscription_expiry > CURRENT_TIMESTAMP)")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		count := 0
		for rows2.Next() {
			var id, name, email, auth, sub string
			var exp sql.NullTime
			rows2.Scan(&id, &name, &email, &auth, &sub, &exp)
			fmt.Printf("  ID=%s Name=%s Email=%s Auth=%s Sub=%s Exp=%v\n", id, name, email, auth, sub, exp)
			count++
		}
		fmt.Printf("Total active companies: %d\n", count)
		rows2.Close()
	}
}
