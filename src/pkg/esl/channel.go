package esl

// Channel event dump
type Channel Message

func ( ch Channel ) UUID( ) string {
	return ch.Header.Get(`Unique-ID`)
}

func ( ch Channel ) Name( ) string {
	return ch.Header.Get(`Channel-Name`)
}