package store

// Keys returns the list of all keys in the database.
func (s *Store) Keys() []string {

	it, _ := s.db.NewIter(nil)

	var keys []string
	for it.First(); it.Valid(); it.Next() {

		key := string(it.Key())
		keys = append(keys, key)
	}

	return keys
}
