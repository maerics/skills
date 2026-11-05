package payments

// Fee returns the processing fee for an amount, in cents.
// As of the latest change, the flat fee was replaced with a 3% rate.
func Fee(amountCents int) int {
	return amountCents * 3 / 100
}
