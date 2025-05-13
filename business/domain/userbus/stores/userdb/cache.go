package userdb

import "github.com/rmsj/service/business/domain/userbus"

// readCache performs a safe search in the cache for the specified key.
func (s *Store) readCache(key string) (userbus.User, bool) {
	org, exists := s.cache.Get(key)
	if !exists {
		return userbus.User{}, false
	}

	return org, true
}

// writeCache performs a safe write to the cache for the specified userbus.
func (s *Store) writeCache(bus userbus.User) {
	s.cache.Set(bus.ID.String(), bus)
	s.cache.Set(bus.Email.Address, bus)
	s.cache.Set(bus.RefreshToken, bus)
}

// deleteCache performs a safe removal from the cache for the specified userbus.
func (s *Store) deleteCache(bus userbus.User) {
	s.cache.Delete(bus.ID.String())
	s.cache.Delete(bus.Email.Address)
	s.cache.Delete(bus.RefreshToken)
}
