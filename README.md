# [Mace](https://github.com/djinn/mace) is a fast caching library for [golang](https://github.com/golang/go)
Mace easy caching mechanism which is sufficiently fast. Mace does not provide
caching over application cluster. For that specific purpose please evaluate
[groupcache](https://github.com/golang/groupcache).

# Key Advantages
  * Fast, because it exists within same process space
  * Does not have expensive JSON serialization, deserialization
  * Supports Native slices, lists, maps, structs and custom types
  * Expiration can be set on each key
  * Ability to setup 'add item' and 'deleted item' events
  * Load item event to allow fetching uncached items

# Mace does not fit where
  * key store is shared in a cluster
  * item specific events need to be generated
  * Where keys are not string
