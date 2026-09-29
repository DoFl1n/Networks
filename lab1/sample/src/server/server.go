package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"lab1/src/proto"
	"math/big"
	"net"
	"strconv"

	log "github.com/mgutz/logxi/v1"
)

// Client - состояние клиента.
type Client struct {
	logger log.Logger    // Объект для печати логов
	conn   *net.TCPConn  // Объект TCP-соединения
	enc    *json.Encoder // Объект для кодирования и отправки сообщений
	matrix []big.Rat     // Текущая матрица
	det    *big.Rat      // Количество полученных от клиента дробей
	dim    int           // Размерность матрицы
}

// NewClient - конструктор клиента, принимает в качестве параметра
// объект TCP-соединения.
func NewClient(conn *net.TCPConn) *Client {
	return &Client{
		logger: log.New(fmt.Sprintf("client %s", conn.RemoteAddr().String())),
		conn:   conn,
		enc:    json.NewEncoder(conn),
		matrix: []big.Rat{},
		det:    big.NewRat(0, 1),
	}
}

// serve - метод, в котором реализован цикл взаимодействия с клиентом.
// Подразумевается, что метод serve будет вызаваться в отдельной go-программе.
func (client *Client) serve() {
	defer client.conn.Close()
	decoder := json.NewDecoder(client.conn)
	for {
		var req proto.Request
		if err := decoder.Decode(&req); err != nil {
			client.logger.Error("cannot decode message", "reason", err)
			break
		} else {
			client.logger.Info("received command", "command", req.Command)
			if client.handleRequest(&req) {
				client.logger.Info("shutting down connection")
				break
			}
		}
	}
}

// handleRequest - метод обработки запроса от клиента. Он возвращает true,
// если клиент передал команду "quit" и хочет завершить общение.
func (client *Client) handleRequest(req *proto.Request) bool {
	switch req.Command {
	case "quit":
		client.respond("ok", nil)
		return true
	case "get_dim":
		client.respond("result", strconv.Itoa(client.dim))
	case "dim":
		errorMsg := ""
		if req.Data == nil {
			errorMsg = "data field is absent"
		} else {
			var jsondim string
			if err := json.Unmarshal(*req.Data, &jsondim); err != nil {
				errorMsg = "malformed data field"
			} else {
				if x, err := strconv.Atoi(jsondim); err != nil {
					errorMsg = "unavailable to convert string, invalid syntax"
				} else {
					client.logger.Info("crafting dimention", "size", jsondim)
					client.dim = x
				}
			}
		}
		if errorMsg == "" {
			client.respond("ok", nil)
		} else {
			client.logger.Error("addition failed", "reason", errorMsg)
			client.respond("failed", errorMsg)
		}
	case "input":
		errorMsg := ""
		if req.Data == nil {
			errorMsg = "data field is absent"
		} else if client.dim <= 0 {
			errorMsg = "unavailable dimention size, please write positive integer"
		} else {
			var matrix proto.Matrix
			if err := json.Unmarshal(*req.Data, &matrix); err != nil {
				errorMsg = "malformed data field"
			} else {
				Matrix := make([]big.Rat, client.dim*client.dim)
				for i, frac := range matrix.Data {
					num := new(big.Int)
					if _, ok := num.SetString(frac.Numerator, 10); !ok {
						errorMsg = "invalid numerator"
						break
					}
					denom := new(big.Int)
					if _, ok := denom.SetString(frac.Denominator, 10); !ok {
						errorMsg = "invalid denominator"
						break
					}
					if denom.Sign() == 0 {
						errorMsg = "division by zero in fraction"
						break
					}
					Matrix[i].SetFrac(num, denom)
				}
				if len(Matrix) != client.dim*client.dim {
					errorMsg = "dimention and size of input are not capable"
				} else {
					client.logger.Info("performing input", "1D matrix", matrix)
					*client.det = det(Matrix)
				}
			}
		}
		if errorMsg == "" {
			client.respond("ok", nil)
		} else {
			client.logger.Error("input failed", "reason", errorMsg)
			client.respond("failed", errorMsg)
		}
	case "det":
		if client.dim <= 0 {
			client.logger.Error("calculation failed", "reason", "matrix doesn't exist")
			client.respond("failed", "matrix doesn't exist")
		} else {
			client.respond("result", &proto.JsonFraction{
				Numerator:   client.det.Num().String(),
				Denominator: client.det.Denom().String(),
			})
		}
	default:
		client.logger.Error("unknown command")
		client.respond("failed", "unknown command")
	}
	return false
}

// respond - вспомогательный метод для передачи ответа с указанным статусом
// и данными. Данные могут быть пустыми (data == nil).
func (client *Client) respond(status string, data interface{}) {
	var raw json.RawMessage
	raw, _ = json.Marshal(data)
	client.enc.Encode(&proto.Response{status, &raw})
}

// swapRows меняет местами строки i и j в матрице размера n x n
func swapRows(i, j, n int, mat []big.Rat) {
	for k := 0; k < n; k++ {
		mat[i*n+k], mat[j*n+k] = mat[j*n+k], mat[i*n+k]
	}
}

// det вычисляет определитель квадратной матрицы n x n методом Гаусса
func det(data []big.Rat) big.Rat {
	n := 0
	for n*n < len(data) {
		n++
	}
	if n*n != len(data) || n == 0 {
		return *new(big.Rat)
	}

	mat := make([]big.Rat, len(data))
	for i := range data {
		mat[i].Set(&data[i])
	}

	sign := 1
	for i := 0; i < n; i++ {
		pivot := -1
		for k := i; k < n; k++ {
			if mat[k*n+i].Sign() != 0 {
				pivot = k
				break
			}
		}
		if pivot == -1 {
			return *new(big.Rat)
		}
		if pivot != i {
			swapRows(i, pivot, n, mat)
			sign = -sign
		}
		for j := i + 1; j < n; j++ {
			if mat[j*n+i].Sign() == 0 {
				continue
			}

			var lambda big.Rat
			lambda.Quo(&mat[j*n+i], &mat[i*n+i])

			var subtrahend big.Rat
			for k := i; k < n; k++ {
				subtrahend.Mul(&lambda, &mat[i*n+k])
				mat[j*n+k].Sub(&mat[j*n+k], &subtrahend)
			}
		}
	}

	var res big.Rat
	res.SetInt64(int64(sign))
	for i := 0; i < n; i++ {
		res.Mul(&res, &mat[i*n+i])
	}

	return res
}

func main() {
	// Работа с командной строкой, в которой может указываться необязательный ключ -addr.
	var addrStr string
	flag.StringVar(&addrStr, "addr", "127.0.0.1:6000", "specify ip address and port")
	flag.Parse()

	// Разбор адреса, строковое представление которого находится в переменной addrStr.
	if addr, err := net.ResolveTCPAddr("tcp", addrStr); err != nil {
		log.Error("address resolution failed", "address", addrStr)
	} else {
		log.Info("resolved TCP address", "address", addr.String())

		// Инициация слушания сети на заданном адресе.
		if listener, err := net.ListenTCP("tcp", addr); err != nil {
			log.Error("listening failed", "reason", err)
		} else {
			// Цикл приёма входящих соединений.
			for {
				if conn, err := listener.AcceptTCP(); err != nil {
					log.Error("cannot accept connection", "reason", err)
				} else {
					log.Info("accepted connection", "address", conn.RemoteAddr().String())

					// Запуск go-программы для обслуживания клиентов.
					go NewClient(conn).serve()
				}
			}
		}
	}
}
