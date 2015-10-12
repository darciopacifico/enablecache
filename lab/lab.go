package main

import (
	"fmt"
	"github.com/golang/groupcache"
)

func main() {

	thumb := groupcache.NewGroup("dlp", 64<<20, groupcache.GetterFunc(
		func(ctx groupcache.Context, key string, dest groupcache.Sink) error {
			fileName := key
			dest.SetBytes(generateThumb(fileName))
			return nil
		}))

	groupcache.RegisterServerStart(func() { fmt.Println("chamando funcao q nao sei pra q eh...") })

	var data []byte
	thumb.Get(nil, "product:123", groupcache.AllocatingByteSliceSink(&data))

	fmt.Println(fmt.Sprintf("resultado retornado pelo get: %v", string(data)))

}

func generateThumb(fileName string) []byte {

	fmt.Println("gerando bytes para serem cacheados")

	return []byte("Um array de bytes para img ... " + fileName)

}
