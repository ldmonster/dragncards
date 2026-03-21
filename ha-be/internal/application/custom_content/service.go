package customcontent
package custom_content

import "github.com/ldmonster/dragncards/ha-be/internal/domain/custom_content"





























}	return s.domain.GetByID(id)func (s *Service) GetByID(id string) (*custom_content.CustomContent, error) {}	return s.domain.Delete(id)func (s *Service) Delete(id string) error {}	return s.domain.Create(pluginID, ownerID, name, data)func (s *Service) Create(pluginID, ownerID, name, data string) (*custom_content.CustomContent, error) {}	return s.domain.ListByOwner(ownerID, pluginID)func (s *Service) ListByOwner(ownerID, pluginID string) ([]*custom_content.CustomContent, error) {}	return s.domain.ListByPlugin(pluginID)func (s *Service) ListByPlugin(pluginID string) ([]*custom_content.CustomContent, error) {}	return &Service{domain: domainSvc}func NewService(domainSvc *custom_content.CustomContentService) *Service {}	domain *custom_content.CustomContentServicetype Service struct {// Service wraps custom_content domain service in application layer.