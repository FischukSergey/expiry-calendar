package seed

const (
	// Демо-логины локального стенда (README).
	adminEmail  = "admin@duekeep.local"
	viewerEmail = "viewer@duekeep.local"
	// Стабильные UUID: admin как в примере GET /me (api-sprint-2).
	adminID  = "11111111-1111-1111-1111-111111111111"
	viewerID = "22222222-2222-2222-2222-222222222222"

	// Демо-пароли локального стенда (README). Не прод-секреты.
	adminPassword  = "admin1234"
	viewerPassword = "viewer1234"
)

// userSeed — строка users. id фиксирован, не gen_random_uuid.
type userSeed struct {
	id       string
	email    string
	password string
	role     string
}

// userSeeds — локальный стенд: admin с каталогом 50+, viewer без шаринга (пустой список).
var userSeeds = []userSeed{
	{id: adminID, email: adminEmail, password: adminPassword, role: "admin"},
	{id: viewerID, email: viewerEmail, password: viewerPassword, role: "viewer"},
}
