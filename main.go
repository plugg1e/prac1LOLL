package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Schema struct {
	Name        string              `json:"name"`
	TuplesLimit int                 `json:"tuples_limit"`
	Structure   map[string][]string `json:"structure"`
}

func main() {
	// Чтение схемы из файла
	file, err := os.Open("scheme.json")
	if err != nil {
		fmt.Println("Ошибка при открытии файла:", err)
		return
	}
	defer file.Close()

	var schema Schema
	if err := json.NewDecoder(file).Decode(&schema); err != nil {
		fmt.Println("Ошибка при декодировании JSON:", err)
		return
	}

	// Интерфейс командной строки
	fmt.Println("Введите команды для управления таблицами (или 'exit' для выхода):")
	var input string
	for {
		fmt.Print("> ")
		// Чтение команды целиком (включая пробелы)
		reader := bufio.NewReader(os.Stdin)
		input, _ = reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if strings.ToLower(input) == "exit" {
			break
		}
		processCommand(schema, input)
	}
}

func processCommand(schema Schema, command string) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		fmt.Println("Пожалуйста, введите команду.")
		return
	}

	switch strings.ToUpper(parts[0]) {
	case "INSERT":
		if len(parts) < 5 {
			fmt.Println("Использование: INSERT таблица значение1, значение2, значение3, значение4")
			return
		}
		tableName := parts[1]
		values := strings.Join(parts[2:], " ")
		insertData(schema, tableName, values)
	case "SELECT":
		if len(parts) == 3 && strings.ToUpper(parts[1]) == "ALL" {
			tableName := parts[2]
			selectAllData(schema, tableName)
		} else if len(parts) < 3 {
			fmt.Println("Использование: SELECT таблица WHERE условие")
			return
		} else {
			tableName := parts[1]
			whereClause := strings.Join(parts[3:], " ")
			selectData(schema, tableName, whereClause)
		}
	case "DELETE":
		if len(parts) < 3 {
			fmt.Println("Использование: DELETE FROM таблица WHERE условие")
			return
		}
		tableName := parts[2]
		whereClause := strings.Join(parts[4:], " ")
		deleteData(schema, tableName, whereClause)
	default:
		fmt.Println("Неизвестная команда.")
	}
}

func insertData(schema Schema, tableName string, values string) {
	if _, exists := schema.Structure[tableName]; !exists {
		fmt.Println("Таблица не существует:", tableName)
		return
	}

	// Создаем или открываем файл в режиме добавления
	file, err := os.OpenFile(tableName+".csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Ошибка при открытии файла:", err)
		return
	}
	defer file.Close()

	// Создаем новый CSV writer
	writer := csv.NewWriter(file)

	// Разбиваем значения по запятой и записываем в файл
	record := strings.Split(values, ", ")
	err = writer.Write(record)
	if err != nil {
		fmt.Println("Ошибка при записи в файл:", err)
		return
	}

	// Сохраняем запись
	writer.Flush()
	if err := writer.Error(); err != nil {
		fmt.Println("Ошибка при сохранении записи:", err)
		return
	}

	fmt.Println("Данные успешно вставлены.")
}

func selectData(schema Schema, tableName string, whereClause string) {
	if _, exists := schema.Structure[tableName]; !exists {
		fmt.Println("Таблица не существует:", tableName)
		return
	}

	file, err := os.Open(tableName + ".csv")
	if err != nil {
		fmt.Println("Ошибка при открытии файла:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Ошибка при чтении файла:", err)
		return
	}

	var found bool
	for _, record := range records {
		for _, field := range record {
			if strings.Contains(field, whereClause) {
				fmt.Println(record)
				found = true
				break
			}
		}
	}

	if !found {
		fmt.Println("Нет записей, соответствующих условию:", whereClause)
	}
}

func selectAllData(schema Schema, tableName string) {
	if _, exists := schema.Structure[tableName]; !exists {
		fmt.Println("Таблица не существует:", tableName)
		return
	}

	file, err := os.Open(tableName + ".csv")
	if err != nil {
		fmt.Println("Ошибка при открытии файла:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Ошибка при чтении файла:", err)
		return
	}

	for _, record := range records {
		fmt.Println(record)
	}
}

func deleteData(schema Schema, tableName string, whereClause string) {
	if _, exists := schema.Structure[tableName]; !exists {
		fmt.Println("Таблица не существует:", tableName)
		return
	}

	file, err := os.ReadFile(tableName + ".csv")
	if err != nil {
		fmt.Println("Ошибка при чтении файла:", err)
		return
	}

	records, err := csv.NewReader(strings.NewReader(string(file))).ReadAll()
	if err != nil {
		fmt.Println("Ошибка при чтении CSV:", err)
		return
	}

	var updatedRecords [][]string
	for _, record := range records {
		found := false
		for _, field := range record {
			if strings.Contains(field, whereClause) {
				found = true
				break
			}
		}
		if !found {
			updatedRecords = append(updatedRecords, record)
		}
	}

	// Записываем обновленные записи обратно в файл
	fileOut, err := os.OpenFile(tableName+".csv", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Ошибка при открытии файла:", err)
		return
	}
	defer fileOut.Close()

	writer := csv.NewWriter(fileOut)
	err = writer.WriteAll(updatedRecords)
	if err != nil {
		fmt.Println("Ошибка при записи в файл:", err)
		return
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		fmt.Println("Ошибка при сохранении записи:", err)
		return
	}

	fmt.Println("Данные успешно удалены.")
}
