package pokedatabases

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"os"

	//Exporatado para Open
	_ "github.com/mattn/go-sqlite3"
)

type bandera int

//Estos errores seran utilizados a lo largo del paquete
var (
	ErrNotPokemons     = errors.New("No tiene pokemons para acceder")
	ErrIncorrectID     = errors.New("ID seleccionada Incorrecta")
	ErrUserExist       = errors.New("El username que usted escogió ya esta en uso")
	ErrPasswordInvalid = errors.New("Password Invalido")
)

//User representa los datos de la tabla users de databases.db
type User struct {
	ID       int
	Username string
	Password string
	Pokemons []Pokemon
}

//Pokemon representa los datos de la tabla pokemons de databases.db
type Pokemon struct {
	ID      int
	Name    string
	Life    int
	Tipo    string
	Level   int
	Ataques []Attack
}

//Attack representa los datos de la tabla attacks de databases.db
type Attack struct {
	ID       int
	Name     string
	Power    int
	Accuracy int
}

//Migracion es una funcion complementaria de Open
//Al ejecutarla se migraran los datos del archivo .sql a un archivo .db
func Migracion() (databases *sql.DB, err error) {
	archivoDB, err := os.Create("databases.db")

	if err != nil {
		return
	}
	archivoDB.Close()

	databases, err = sql.Open("sqlite3", "./databases.db?_foreign_keys=on")

	if err != nil {
		return
	}

	archivoSQL, err := os.Open("databases.sql")

	if err != nil {
		return
	}

	defer archivoSQL.Close()

	contendio, err := ioutil.ReadAll(archivoSQL)

	if err != nil {
		return
	}

	_, err = databases.Exec(string(contendio))
	if err != nil {
		return
	}

	return
}

//Open Abrira el archivo .db o en su defecto ejecutara Migracion
func Open() (databases *sql.DB, err error) {

	archivo, err := os.Open("databases.db")

	if err != nil {
		if os.IsNotExist(err) {

			databases, err := Migracion()

			if err != nil {
				return databases, err
			}

			return databases, err
		}
		return
	}
	defer archivo.Close()

	databases, err = sql.Open("sqlite3", "./databases.db?_foreign_keys=on")

	if err != nil {
		return
	}

	return
}

//GetUser vericara si existe un user registrado con los parametros predefinidos
func GetUser(databases *sql.DB, usernameScan, passwordScan string) (user *User, err error) {

	var userAux User

	row := databases.QueryRow("SELECT id,username,password FROM users WHERE username = ? AND password = ?", usernameScan, passwordScan)

	err = row.Scan(&userAux.ID, &userAux.Username, &userAux.Password)

	if err != nil {
		return
	}

	user = &userAux

	return

}

//CheckIfUserAlreadyExist verifica si ya existe un usuario registrado con el mismo username
func CheckIfUserAlreadyExist(databases *sql.DB, usernameScan string) (check bool, err error) {

	var id int

	row := databases.QueryRow("SELECT id FROM users WHERE username = ? ", usernameScan)

	err = row.Scan(&id)

	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		}
		check = true
		return
	}
	check = true
	return
}

//InsertUser inserta a la base de datos el nuevo usuario
func InsertUser(databases *sql.DB, usernameScan, passwordScan string) (err error) {

	stmt, err := databases.Prepare("INSERT INTO users(username, password) VALUES(?,?)")

	if err != nil {
		return
	}
	_, err = stmt.Exec(usernameScan, passwordScan)

	if err != nil {

		return
	}

	return
}

//GetPokemonsWithUserID da como retorno un slice de pokemon
func GetPokemonsWithUserID(databases *sql.DB, id int) (pokes []Pokemon, err error) {

	rows, err := databases.Query("SELECT DISTINCT pokemons.id,pokemons.name,pokemons.life,pokemons.type,pokemons.level FROM users_pokemons INNER JOIN users ON users_pokemons.user_id = users.id INNER JOIN pokemons ON users_pokemons.poke_id = pokemons.id WHERE users.id = ?", id)

	if err != nil {
		return pokes, err
	}

	defer rows.Close()

	for rows.Next() {
		var newPokemon Pokemon

		err = rows.Scan(&newPokemon.ID, &newPokemon.Name, &newPokemon.Life, &newPokemon.Tipo, &newPokemon.Level)

		if err != nil {
			return
		}

		err = PutAtaques(databases, &newPokemon)

		if err != nil {
			return
		}

		pokes = append(pokes, newPokemon)

	}

	return
}

//GetPokemonWithIDAndUserID como retorno un slice de pokemon
func GetPokemonWithIDAndUserID(databases *sql.DB, idPoke, idUser int) (poke *Pokemon, err error) {

	var pokeAux Pokemon

	row := databases.QueryRow("SELECT DISTINCT pokemons.id,pokemons.name,pokemons.life,pokemons.type,pokemons.level FROM users_pokemons INNER JOIN pokemons ON users_pokemons.poke_id = pokemons.id WHERE users_pokemons.poke_id = ? AND users_pokemons.user_id = ?", idPoke, idUser)

	if err != nil {
		return
	}

	err = row.Scan(&pokeAux.ID, &pokeAux.Name, &pokeAux.Life, &pokeAux.Tipo, &pokeAux.Level)

	if err != nil {

		return
	}

	err = PutAtaques(databases, &pokeAux)

	if err != nil {
		return
	}

	poke = &pokeAux
	return
}

//GetPokemonWithID como retorno un slice de pokemon
func GetPokemonWithID(databases *sql.DB, id int) (poke *Pokemon, err error) {

	var pokeAux Pokemon

	row := databases.QueryRow("SELECT DISTINCT pokemons.id,pokemons.name,pokemons.life,pokemons.type,pokemons.level FROM pokemons WHERE pokemons.id = ?", id)

	if err != nil {
		return
	}

	err = row.Scan(&pokeAux.ID, &pokeAux.Name, &pokeAux.Life, &pokeAux.Tipo, &pokeAux.Level)

	if err != nil {
		return
	}

	err = PutAtaques(databases, &pokeAux)

	if err != nil {
		return
	}

	poke = &pokeAux

	return
}

//GetAllPokemons como retorno un slice de pokemon
func GetAllPokemons(databases *sql.DB) (pokes []Pokemon, err error) {

	rows, err := databases.Query("SELECT DISTINCT pokemons.id,pokemons.name,pokemons.life,pokemons.type,pokemons.level FROM pokemons")

	if err != nil {
		return pokes, err
	}

	defer rows.Close()

	for rows.Next() {
		var newPokemon Pokemon

		err = rows.Scan(&newPokemon.ID, &newPokemon.Name, &newPokemon.Life, &newPokemon.Tipo, &newPokemon.Level)

		if err != nil {
			return pokes, err
		}

		err = PutAtaques(databases, &newPokemon)

		if err != nil {
			return pokes, err
		}

		pokes = append(pokes, newPokemon)

	}
	return

}

//PutAtaques complementa a SelecionarPokemons dandole los ataques
func PutAtaques(databases *sql.DB, newPokemon *Pokemon) (err error) {

	rows, err := databases.Query("SELECT DISTINCT attacks.id,attacks.name,attacks.power,attacks.accuracy FROM pokemons_attacks INNER JOIN attacks ON pokemons_attacks.attack_id=attacks.id WHERE pokemons_attacks.poke_id=?", newPokemon.ID)

	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {

		var newAttack Attack

		err = rows.Scan(&newAttack.ID, &newAttack.Name, &newAttack.Power, &newAttack.Accuracy)

		if err != nil {
			return
		}

		newPokemon.Ataques = append(newPokemon.Ataques, newAttack)
	}

	return
}

//InsertarNuevoPoke inserta un nuevo pokemon
func InsertarNuevoPoke(tx *sql.Tx, newPokemon Pokemon) (err error) {

	_, err = tx.Exec("INSERT INTO pokemons(name,life,type,level) VALUES (?,?,?,?)", newPokemon.Name, newPokemon.Life, newPokemon.Tipo, newPokemon.Level)

	if err != nil {
		return
	}

	return
}

//InsertarRelacionNuevoPoke inserta datos en la tabla pivote users_pokemons
func InsertarRelacionNuevoPoke(tx *sql.Tx, user User, newPokemon Pokemon, idAux int) (err error) {
	_, err = tx.Exec("INSERT INTO users_pokemons (user_id,poke_id) VALUES (?,?)", user.ID, idAux)

	if err != nil {
		return
	}

	return
}

//InsertarRelacionAtaques inserta datos en la tabla pivote pokemons_attacks
func InsertarRelacionAtaques(tx *sql.Tx, attacks []Attack, idAux int) (err error) {

	for i := range attacks {

		_, err = tx.Exec("INSERT INTO pokemons_attacks (poke_id,attack_id) VALUES (?,?)", idAux, attacks[i].ID)

		if err != nil {
			return err
		}

	}
	return
}

//DeletePoke borra un pokemon
func DeletePoke(databases *sql.DB, user User, poke Pokemon) (err error) {

	stmtDelete, err := databases.Prepare("DELETE FROM users_pokemons WHERE users_pokemons.poke_id = ?;")

	if err != nil {
		return err
	}

	_, err = stmtDelete.Exec(poke.ID)

	if err != nil {
		return err
	}

	return
}

//BorrarCuenta te permite eliminar tu cuenta
func BorrarCuenta(databases *sql.DB, userID int, userPassword, preguntaSeguridad string) (check bool, err error) {

	if preguntaSeguridad == userPassword {

		stmtDelete, err := databases.Prepare("DELETE FROM users where id = ?")

		if err != nil {
			return check, err
		}

		_, err = stmtDelete.Exec(userID)

		if err != nil {
			return check, err
		}

		check = true
		return check, err

	}
	err = ErrPasswordInvalid
	return

}
