### Memcache

Usage

```bash
cache := memcache.New(5*time.Minute, 10*time.Minute)

cache.Add("first-value", "some value", 5*time.Minute)

cache.IsExist("some key")

data, err := cache.Get("third")
if err != nil {
    fmt.Println(err)
}

cache.Delete("test key")
```