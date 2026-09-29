package matrix

import (
	"fmt"
)

type Fraction struct {
	num int
	dem int
}

func add_Fractions(a, b Fraction) Fraction {
	var c Fraction
	c.num = a.num*b.dem + b.num*a.dem
	c.dem = a.dem * b.dem
	return normalize_Fraction(c)
}

func sub_Fractions(a, b Fraction) Fraction {
	var c Fraction
	c.num = a.num*b.dem - b.num*a.dem
	c.dem = a.dem * b.dem
	return normalize_Fraction(c)
}

func mul_Fractions(a, b Fraction) Fraction {
	var c Fraction
	c.num = a.num * b.num
	c.dem = a.dem * b.dem
	return normalize_Fraction(c)
}

func dev_Fractions(a, b Fraction) Fraction {
	var c Fraction
	c.num = a.num * b.dem
	c.dem = a.dem * b.num
	return normalize_Fraction(c)
}

func gcd(a, b int) int {
	if b == 0 {
		return a
	}
	return gcd(b, a%b)
}

func normalize_Fraction(a Fraction) Fraction {
	g := gcd(a.num, a.dem)
	a.num /= g
	a.dem /= g
	if a.dem < 0 {
		a.num = -a.num
		a.dem = -a.dem
	}
	return a
}

func print_Fraction(a Fraction) {
	fmt.Printf("%d/%d ", a.num, a.dem)
}

func swap_Rows(i, j, M int, matrix []Fraction) {
	for k := 0; k < M; k++ {
		matrix[i*M+k], matrix[j*M+k] = matrix[j*M+k], matrix[i*M+k]
	}
}

func det(data []Fraction) Fraction {
	n := 0
	for n*n < len(data) {
		n++
	}
	if n*n != len(data) || n == 0 {
		return Fraction{num: 0, dem: 1}
	}

	mat := make([]Fraction, len(data))
	copy(mat, data)

	sign := 1

	for i := 0; i < n; i++ {
		pivot := -1
		for k := i; k < n; k++ {
			if mat[k*n+i].num != 0 {
				pivot = k
				break
			}
		}

		if pivot == -1 {
			return Fraction{num: 0, dem: 1}
		}

		if pivot != i {
			swap_Rows(i, pivot, n, mat)
			sign = -sign
		}

		for j := i + 1; j < n; j++ {
			if mat[j*n+i].num == 0 {
				continue
			}
			lambda := dev_Fractions(mat[j*n+i], mat[i*n+i])
			for k := i; k < n; k++ {
				subtrahend := mul_Fractions(lambda, mat[i*n+k])
				mat[j*n+k] = sub_Fractions(mat[j*n+k], subtrahend)
			}
		}
	}

	res := Fraction{num: sign, dem: 1}
	for i := 0; i < n; i++ {
		res = mul_Fractions(res, mat[i*n+i])
	}

	return normalize_Fraction(res)
}
