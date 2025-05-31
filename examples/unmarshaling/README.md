# Unmarshaling Example

Shows how to unmarshal data with validation and default value application.

## What it shows

- `Unmarshal()` method validates and unmarshals data
- Default values from schema are automatically applied
- Validation errors during unmarshal
- Unmarshal to structs or maps

## Key Features

- Validates data before unmarshaling
- Applies schema defaults to missing fields
- Returns validation errors if data is invalid
- Works with JSON, maps, and structs as input

## Run

```bash
go run main.go
```

## Output

```
Unmarshaling with Defaults
==========================
1. JSON with missing optional fields:
   Result: {Name:Alice Age:25 Country:US Active:true}

2. Map with some values provided:
   Result: {Name:Bob Age:30 Country:Canada Active:true}

3. Validation failure during unmarshal:
   Error: validation failed: name: [Value should be at least 1 characters long], age: [Value should be at least 0]

4. Unmarshal to map:
   Result: map[active:true age:35 country:US name:Charlie]
``` 
