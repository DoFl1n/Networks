package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"lab1/src/proto"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/skorobogatov/input"
)

// interact - функция, содержащая цикл взаимодействия с сервером.
func interact(conn *net.TCPConn) {
	defer conn.Close()
	encoder, decoder := json.NewEncoder(conn), json.NewDecoder(conn)
	for {
		// Чтение команды из стандартного потока ввода
		fmt.Printf("command = ")
		command := input.Gets()

		// Отправка запроса.
		switch command {
		case "quit":
			send_request(encoder, "quit", nil)
			return
		case "dim":
			var str string
			fmt.Printf("dimentional size =")
			str = input.Gets()
			send_request(encoder, "dim", &str)
		case "input":
			// Здесь мы должны получить с сервера размерность, которая на нем сохранена
			send_request(encoder, "get_dim", nil)
			var resp proto.Response
			if err := decoder.Decode(&resp); err != nil {
				fmt.Printf("error: %v\n", err)
				break
			}
			switch resp.Status {
			case "result":
				// Обрабатываем статус
				if resp.Data == nil {
					fmt.Printf("error: data field is absent in response\n")
				} else {
					var x string
					if err := json.Unmarshal(*resp.Data, &x); err != nil {
						fmt.Printf("error0: malformed data field in response\n")
					} else {
						if y, err := strconv.Atoi(x); err != nil {
							fmt.Printf("error: unvailable dimention size")
						} else {
							if y <= 0 {
								fmt.Printf("error: unvailable dimention size")
							} else {
								fmt.Printf("input massive with a size of %v (пример ввода: 1/2, 45/7, ... , 7/9 строки можно переносить)", y*y)
								Matrix := parseMatrixInput(y)
								send_request(encoder, "input", &Matrix)
							}
						}
					}
				}
			}
		case "det":
			send_request(encoder, "det", nil)
		default:
			fmt.Printf("error: unknown command\n")
			continue
		}

		// Получение ответа.
		var resp proto.Response
		if err := decoder.Decode(&resp); err != nil {
			fmt.Printf("error: %v\n", err)
			break
		}

		// Вывод ответа в стандартный поток вывода.
		switch resp.Status {
		case "ok":
			fmt.Printf("ok\n")
		case "failed":
			if resp.Data == nil {
				fmt.Printf("error: data field is absent in response\n")
			} else {
				var errorMsg string
				if err := json.Unmarshal(*resp.Data, &errorMsg); err != nil {
					fmt.Printf("error: malformed data field in response\n")
				} else {
					fmt.Printf("failed: %s\n", errorMsg)
				}
			}
		case "result":
			if resp.Data == nil {
				fmt.Printf("error: data field is absent in response\n")
			} else {
				var frac proto.JsonFraction
				if err := json.Unmarshal(*resp.Data, &frac); err != nil {
					fmt.Printf("error: malformed data field in response\n")
				} else {
					fmt.Printf("result: %s/%s\n", frac.Numerator, frac.Denominator)
				}
			}
		default:
			fmt.Printf("error: server reports unknown status %q\n", resp.Status)
		}
	}
}

// send_request - вспомогательная функция для передачи запроса с указанной командой
// и данными. Данные могут быть пустыми (data == nil).
func send_request(encoder *json.Encoder, command string, data interface{}) {
	var raw json.RawMessage
	raw, _ = json.Marshal(data)
	encoder.Encode(&proto.Request{command, &raw})
}

var fracRegex = regexp.MustCompile(`([+-]?\d+)(?:\s*/\s*([+-]?\d+))?`)

func parseMatrixInput(y int) proto.Matrix {
	total := y * y
	var Matrix proto.Matrix

	fmt.Printf("Введите элементы матрицы (всего %d шт., пример: 1/2, 45/7, -3, 2 / 3; строки можно переносить):\n", total)

	for len(Matrix.Data) < total {
		str := input.Gets()

		// Ищем все вхождения чисел и дробей в текущей строке
		matches := fracRegex.FindAllStringSubmatch(str, -1)

		for _, match := range matches {
			if len(Matrix.Data) >= total {
				break
			}

			numStr := match[1]
			denomStr := match[2]

			// Обработка случая, когда ввели просто число без слэша (например, "5" -> "5/1")
			if denomStr == "" {
				denomStr = "1"
			}

			// Проверка деления на 0
			if denomStr == "0" || denomStr == "-0" || denomStr == "+0" {
				fmt.Println("Ошибка: знаменатель не может быть равен 0. Дробь пропущена.")
				continue
			}

			// Нормализация знаков: переносим минус из знаменателя в числитель (например, 1/-2 -> -1/2)
			if strings.HasPrefix(denomStr, "-") {
				denomStr = denomStr[1:]
				if strings.HasPrefix(numStr, "-") {
					numStr = numStr[1:]
				} else if strings.HasPrefix(numStr, "+") {
					numStr = "-" + numStr[1:]
				} else {
					numStr = "-" + numStr
				}
			} else if strings.HasPrefix(denomStr, "+") {
				denomStr = denomStr[1:]
			}
			if strings.HasPrefix(numStr, "+") {
				numStr = numStr[1:]
			}

			// Заполнение структуры дроби.
			// ВАРИАНТ А: Если в proto.Fraction поля строковые (Numerator/Denominator или Num/Denom):
			frac := proto.JsonFraction{
				Numerator:   numStr,
				Denominator: denomStr,
			}

			// ВАРИАНТ Б: Если в вашей proto.Fraction поля типа int/int64:
			// n, _ := strconv.Atoi(numStr)
			// d, _ := strconv.Atoi(denomStr)
			// frac := proto.Fraction{Num: n, Dem: d}

			Matrix.Data = append(Matrix.Data, frac)
			fmt.Printf("Добавлено: %s/%s | Осталось ввести: %d\n", numStr, denomStr, total-len(Matrix.Data))
		}
	}

	return Matrix
}

func main() {
	// Работа с командной строкой, в которой может указываться необязательный ключ -addr.
	var addrStr string
	flag.StringVar(&addrStr, "addr", "127.0.0.1:6000", "specify ip address and port")
	flag.Parse()

	// Разбор адреса, установка соединения с сервером и
	// запуск цикла взаимодействия с сервером.
	if addr, err := net.ResolveTCPAddr("tcp", addrStr); err != nil {
		fmt.Printf("error: %v\n", err)
	} else if conn, err := net.DialTCP("tcp", nil, addr); err != nil {
		fmt.Printf("error: %v\n", err)
	} else {
		interact(conn)
	}
}
