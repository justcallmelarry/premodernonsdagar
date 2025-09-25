package aggregation

func AggregateStats() error {
	err := aggregatePlayerStats()
	if err != nil {
		return err
	}

	err = UpdateEvents()
	if err != nil {
		return err
	}

	return nil
}
