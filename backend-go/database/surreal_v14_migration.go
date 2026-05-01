package database

// SurrealDB v1.4.0 API Migration Guide
// 
// Breaking Changes:
// 1. surrealdb.Unmarshal() removed -> Use surrealdb.SmartUnmarshal() or JSON marshaling
// 2. db.Query() now returns *[]surrealdb.QueryResult[T] instead of interface{}
// 3. Functional API introduced: surrealdb.Query[T](ctx, db, sql, params)
//
// Migration Strategy:
// - For raw queries: Use functional API surrealdb.Query[interface{}](ctx, db, sql, params)
// - For unmarshaling: Use surrealdb.SmartUnmarshal() or JSON marshal/unmarshal pattern
// - For typed queries: Use surrealdb.Select[T](), Create[T](), Update[T](), Delete[T]()

// ============================================================================
// Method 1: SmartUnmarshal (Direct Replacement)
// ============================================================================
// If you have raw query results and need to map to a struct:
//
//   res, err := db.Query(sql, vars)
//   if err != nil {
//       return err
//   }
//   
//   var users []User
//   ok, err := surrealdb.SmartUnmarshal(res, &users)
//   if err != nil {
//       return err
//   }
//
// Note: SmartUnmarshal returns (bool, error). The bool indicates success.

// ============================================================================
// Method 2: Functional API with Type Parameters (Recommended)
// ============================================================================
// Use the functional API for type-safe queries:
//
//   results, err := surrealdb.Query[interface{}](ctx, db, sql, params)
//   if err != nil {
//       return err
//   }
//   
//   if len(*results) > 0 {
//       firstResult := (*results)[0].Result
//       // Handle firstResult based on your needs
//   }

// ============================================================================
// Method 3: JSON Marshal/Unmarshal Pattern (Current Implementation)
// ============================================================================
// This is what surreal.go currently uses in queryAndUnmarshal():
//
//   results, err := surrealdb.Query[interface{}](ctx, db, sql, params)
//   if err != nil {
//       return err
//   }
//   
//   if len(*results) == 0 {
//       return fmt.Errorf("no results")
//   }
//   
//   firstResult := (*results)[0]
//   jsonData, err := json.Marshal(firstResult.Result)
//   if err != nil {
//       return err
//   }
//   
//   var target YourStruct
//   if err := json.Unmarshal(jsonData, &target); err != nil {
//       return err
//   }

// ============================================================================
// Method 4: High-Level Typed Methods (Simplest)
// ============================================================================
// For simple CRUD operations, use the high-level methods:
//
//   // Select
//   users, err := surrealdb.Select[User](ctx, db, "user")
//   
//   // Create
//   created, err := surrealdb.Create[User](ctx, db, "user", userData)
//   
//   // Update
//   updated, err := surrealdb.Update[User](ctx, db, "user:123", userData)
//   
//   // Delete
//   deleted, err := surrealdb.Delete[User](ctx, db, "user:123")

// ============================================================================
// Common Errors and Solutions
// ============================================================================
//
// Error: "undefined: surrealdb.Unmarshal"
// Solution: Replace with surrealdb.SmartUnmarshal() or use JSON marshaling
//
// Error: "cannot use result (type interface{}) as type *[]surrealdb.QueryResult"
// Solution: Use functional API: surrealdb.Query[interface{}](ctx, db, sql, params)
//
// Error: "too many arguments in call to db.Query"
// Solution: Use functional API or ensure you're using the correct method signature

// ============================================================================
// Migration Checklist
// ============================================================================
// [ ] Replace all surrealdb.Unmarshal() calls with SmartUnmarshal()
// [ ] Update db.Query() calls to use functional API
// [ ] Add context parameter to all database operations
// [ ] Consider using high-level typed methods for simple CRUD
// [ ] Test all query result unmarshaling logic
// [ ] Update error handling for new return types
