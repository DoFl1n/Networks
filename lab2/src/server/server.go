package main

import (
	"fmt"      // пакет для форматированного ввода вывода
	"log"      // пакет для логирования
	"net/http" // пакет для поддержки HTTP протокола
	"strings"  // пакет для работы с  UTF-8 строками

	"github.com/SlyMarbo/rss"
)

func HomeRouterHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()       // анализ аргументов,
	fmt.Println(r.Form) // ввод информации о форме на стороне сервера
	fmt.Println("path", r.URL.Path)
	fmt.Println("scheme", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])

	rssObject, err := rss.Fetch("https://news.mail.ru/rss/90/")

	if err != nil {
		fmt.Fprintln(w, "Ошибка при загрузке RSS:", err)
		return
	}
	for v := range rssObject.Items {
		item := rssObject.Items[v]

		fmt.Fprintf(w, "<a href=\"%v\">%v</a><br><br>", item.Link, item.Title)
		fmt.Fprintf(w, "Описание: %v<br><br>", item.Summary)
	}
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
}

func main() {
	http.HandleFunc("/", HomeRouterHandler)  // установим роутер
	err := http.ListenAndServe(":9000", nil) // задаем слушать порт
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
