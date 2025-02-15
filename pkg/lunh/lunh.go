package lunh

func Validate(number string) bool {
	return lunh([]byte(number))
}

func lunh(s []byte) bool {
	n := len(s)
	sum := 0
	parity := n % 2
	for i := range n - 1 {
		digit := int(s[i] - '0')
		if i%2 != parity {
			sum += digit
		} else if digit > 4 {
			sum += 2*digit - 9
		} else {
			sum += 2 * digit
		}
	}
	lastDigit := int(s[n-1] - '0')
	return lastDigit == (10-sum%10)%10
}
