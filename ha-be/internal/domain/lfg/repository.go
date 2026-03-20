package lfg

type LfgRepository interface {
	Create(post *LfgPost) error
	List(pluginID string) ([]*LfgPost, error)
	Delete(id string) error
}
