package utils

func TestMinUnit(t *testing.T) {
	val := MinUnit(0.11111, 0.001)
	log.Printf("val: %v", val)
	time.Sleep(time.Second)
}
