# [Mace](https://github.com/djinn/mace)
## Fast caching data objects fast within [golang](https://github.com/golang/go) application 

[![GoDoc](https://godoc.org/github.com/djinn/mace?status.svg)](https://godoc.org/github.com/djinn/mace)

Mace saves time setting up separate server. It is fast because its in-process and does not have serialization, deserialization costs. 

## How to install 

      go get github.com/djinn/mace

## Example
  ```go
  package main

  import (
    "fmt"
    "time"

    "github.com/djinn/mace"
  )

  // Key in mace is always string type
  // Value can be declared to be of arbitrary type
  // Keys & values in cache2go can be off arbitrary types, e.g. a struct.

  type ProductType struct {
    ProductId   int64
    ProductName string
    Variants    string //declared string to keep it simple
    Inventory   []uint
  }

  func main() {
    //Declare cache bucket
    cache := mace.Mace("product")

    // Declare cache object with alive value. Lets say 10 seconds
    product := ProductType{
      522013,
      "Nike Flyknit",
      "black and blue",
      []uint{1, 2, 3},
    }
    cache.Set("522013", &product, 5*time.Millisecond)

    // Let's retrieve the item from the cache.
    res, err := cache.Get("522013")
    if err == nil {
      fmt.Println("Found value in key:", res.Data().(*ProductType))
    } else {
      fmt.Println("Error retrieving value from cache:", err)
    }

    // Wait for the item to expire in cache.
    time.Sleep(6 * time.Second)
    res, err = cache.Get("522013")
    if err != nil {
      fmt.Println("Item is not cached (anymore).")
    }
    val := "string"
    // Add another item that never expires.
    cache.Set("471983", &val, 0)

    // mace supports a few handy callbacks and loading mechanisms.
    cache.SetOnDeleteItem(func(e *mace.MaceItem) {
      fmt.Println("Deleting:", e.Key(), *e.Data().(*string))
    })

    // Remove the item from the cache.
    cache.Delete("471983")

    // And wipe the entire cache table.
    cache.Flush()
  }
  ```






## Key Advantages
  * Fast, because it exists within same process space
  * Does not have expensive JSON serialization, deserialization
  * Supports Native slices, lists, maps, structs and custom types
  * Expiration can be set on each key
  * Ability to setup 'add item' and 'deleted item' events
  * Load item event to allow fetching uncached items

## Mace does not fit where
  * key store is shared in a cluster
  * item specific events need to be generated
  * Where keys are not string


