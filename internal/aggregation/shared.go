package aggregation

func AggregateStats() error {
	err := aggregatePlayerStats()
	if err != nil {
		return err
	}

	err = generateEventsList()
	if err != nil {
		return err
	}

	err = generateLeaderboards()
	if err != nil {
		return err
	}

	return nil
}
