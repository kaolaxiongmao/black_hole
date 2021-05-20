package utils

func TestLark(t *testing.T) {
	ctx := context.Background()
	SendLarkNoticeToDefaultUser(ctx, "fff")
}

