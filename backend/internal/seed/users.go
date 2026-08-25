package seed

const (
	adminEmail  = "admin@duekeep.local"
	viewerEmail = "viewer@duekeep.local"
	adminID     = "11111111-1111-1111-1111-111111111111"
	viewerID    = "22222222-2222-2222-2222-222222222222"

	// Демо-пароли локального стенда (README). Не прод-секреты.
	adminPassword  = "admin1234"
	viewerPassword = "viewer1234"
)

type userSeed struct {
	id       string
	email    string
	password string
	role     string
}

var userSeeds = []userSeed{
	{id: adminID, email: adminEmail, password: adminPassword, role: "admin"},
	{id: viewerID, email: viewerEmail, password: viewerPassword, role: "viewer"},
}
