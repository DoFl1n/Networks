package main

import (
	"fmt"

	"github.com/SlyMarbo/rss"
)

func main() {

	rssObject, err := rss.Fetch("http://somehost.ru/rss_vk.xml")

	if err != nil {
		fmt.Println("Ошибка при загрузке RSS:", err)
		return
	}
	for v := range rssObject.Items {

		item := rssObject.Items[v]

		fmt.Println()

		fmt.Printf("Item Number : %d\n", v)

		fmt.Printf("Title       : %s\n", item.Title)

		fmt.Printf("Link        : %s\n", item.Link)

		fmt.Printf("Description : %s\n", item.Description)

		fmt.Printf("Guid        : %s\n", item.Guid.Value)

	}

}
