package DynService

type Service interface {
	Update(ip string) (err error)
}
