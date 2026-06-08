package service

import (
	"cmp"
	"slices"
	"strings"
	"unicode/utf8"

	"fms-project/internal/domain/entity"
)

type CategoryMatcherServiceConfig struct{}

type CategoryMatcherService struct {
	dictionary []dictionaryEntry
}

type dictionaryEntry struct {
	keyword string
	entry   categoryEntry
}

type categoryEntry struct {
	name   string
	weight int
}

func NewCategoryMatcherService(cfg *CategoryMatcherServiceConfig) *CategoryMatcherService {
	return &CategoryMatcherService{
		dictionary: buildDictionary(),
	}
}

func (s *CategoryMatcherService) Match(items []entity.DraftItem, categories []entity.Category) []entity.DraftItem {
	if len(categories) == 0 {
		return items
	}

	catMap := make(map[string]entity.Category)
	for _, c := range categories {
		catMap[c.Name] = c
	}

	result := make([]entity.DraftItem, len(items))
	for i, item := range items {
		result[i] = item

		nameLower := strings.ToLower(item.Name)
		nameNormalized := normalizeName(nameLower)

		bestCategory := ""
		bestScore := 0

		for _, dictEntry := range s.dictionary {
			score := calculateScore(nameLower, nameNormalized, dictEntry.keyword, dictEntry.entry.weight)
			if score > bestScore {
				bestScore = score
				bestCategory = dictEntry.entry.name
			}
		}

		if bestScore > 0 {
			if cat, ok := catMap[bestCategory]; ok {
				result[i].Category = cat
			}
		}
	}

	return result
}

func calculateScore(nameLower, nameNormalized, keyword string, weight int) int {
	if strings.Contains(nameLower, keyword) {
		return weight
	}

	normalizedKeyword := normalizeName(keyword)
	if strings.Contains(nameNormalized, normalizedKeyword) {
		return weight
	}

	return 0
}

func normalizeName(s string) string {
	replacements := map[string]string{
		"а": "a", "б": "b", "в": "v", "г": "g", "д": "d", "е": "e", "ё": "e",
		"ж": "zh", "з": "z", "и": "i", "й": "y", "к": "k", "л": "l", "м": "m",
		"н": "n", "о": "o", "п": "p", "р": "r", "с": "s", "т": "t", "у": "u",
		"ф": "f", "х": "h", "ц": "ts", "ч": "ch", "ш": "sh", "щ": "sch",
		"ъ": "", "ы": "y", "ь": "", "э": "e", "ю": "yu", "я": "ya",
	}

	var result strings.Builder
	for _, r := range s {
		if repl, ok := replacements[string(r)]; ok {
			result.WriteString(repl)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func buildDictionary() []dictionaryEntry {
	raw := map[string]categoryEntry{
		"хлеб":             {name: "Продукты", weight: 10},
		"батон":            {name: "Продукты", weight: 10},
		"булк":             {name: "Продукты", weight: 10},
		"молок":            {name: "Продукты", weight: 10},
		"кефир":            {name: "Продукты", weight: 10},
		"йогурт":           {name: "Продукты", weight: 10},
		"сыр":              {name: "Продукты", weight: 10},
		"творог":           {name: "Продукты", weight: 10},
		"творож":           {name: "Продукты", weight: 10},
		"масло":            {name: "Продукты", weight: 10},
		"яйц":              {name: "Продукты", weight: 10},
		"мяс":              {name: "Продукты", weight: 10},
		"куриц":            {name: "Продукты", weight: 10},
		"говядин":          {name: "Продукты", weight: 10},
		"свинин":           {name: "Продукты", weight: 10},
		"фарш":             {name: "Продукты", weight: 10},
		"колбас":           {name: "Продукты", weight: 10},
		"сосиск":           {name: "Продукты", weight: 10},
		"рыб":              {name: "Продукты", weight: 10},
		"лосос":            {name: "Продукты", weight: 10},
		"тунец":            {name: "Продукты", weight: 10},
		"овощ":             {name: "Продукты", weight: 10},
		"фрукт":            {name: "Продукты", weight: 10},
		"яблок":            {name: "Продукты", weight: 10},
		"банан":            {name: "Продукты", weight: 10},
		"картоф":           {name: "Продукты", weight: 10},
		"морков":           {name: "Продукты", weight: 10},
		"лук":              {name: "Продукты", weight: 10},
		"помид":            {name: "Продукты", weight: 10},
		"огурц":            {name: "Продукты", weight: 10},
		"круп":             {name: "Продукты", weight: 10},
		"рис":              {name: "Продукты", weight: 10},
		"греч":             {name: "Продукты", weight: 10},
		"макарон":          {name: "Продукты", weight: 10},
		"сахар":            {name: "Продукты", weight: 10},
		"соль":             {name: "Продукты", weight: 10},
		"мук":              {name: "Продукты", weight: 10},
		"чай":              {name: "Продукты", weight: 10},
		"кофе":             {name: "Продукты", weight: 10},
		"сок":              {name: "Продукты", weight: 10},
		"вода":             {name: "Продукты", weight: 10},
		"напиток":          {name: "Продукты", weight: 10},
		"печень":           {name: "Продукты", weight: 10},
		"конфет":           {name: "Продукты", weight: 10},
		"шоколад":          {name: "Продукты", weight: 10},
		"паста тв":         {name: "Продукты", weight: 10},
		"паста том":        {name: "Продукты", weight: 10},
		"шампун":           {name: "Гигиена и уход", weight: 10},
		"гель душ":         {name: "Гигиена и уход", weight: 10},
		"гель для душа":    {name: "Гигиена и уход", weight: 10},
		"мыл":              {name: "Гигиена и уход", weight: 10},
		"зубн":             {name: "Гигиена и уход", weight: 10},
		"паст":             {name: "Гигиена и уход", weight: 10},
		"щетк":             {name: "Гигиена и уход", weight: 10},
		"дезодорант":       {name: "Гигиена и уход", weight: 10},
		"бритв":            {name: "Гигиена и уход", weight: 10},
		"станок":           {name: "Гигиена и уход", weight: 10},
		"крем":             {name: "Гигиена и уход", weight: 10},
		"лосьон":           {name: "Гигиена и уход", weight: 10},
		"маск":             {name: "Гигиена и уход", weight: 10},
		"бальзам":          {name: "Гигиена и уход", weight: 10},
		"прокладк":         {name: "Гигиена и уход", weight: 10},
		"тампон":           {name: "Гигиена и уход", weight: 10},
		"подгуз":           {name: "Гигиена и уход", weight: 10},
		"салфетка влажн":   {name: "Гигиена и уход", weight: 10},
		"ват":              {name: "Гигиена и уход", weight: 10},
		"палочк":           {name: "Гигиена и уход", weight: 10},
		"диск":             {name: "Гигиена и уход", weight: 10},
		"туалетн бумаг":    {name: "Гигиена и уход", weight: 10},
		"посуда":           {name: "Товары для дома", weight: 10},
		"тарелк":           {name: "Товары для дома", weight: 10},
		"кружк":            {name: "Товары для дома", weight: 10},
		"стакан":           {name: "Товары для дома", weight: 10},
		"сковород":         {name: "Товары для дома", weight: 10},
		"кастрюл":          {name: "Товары для дома", weight: 10},
		"контейнер":        {name: "Товары для дома", weight: 10},
		"пакет":            {name: "Товары для дома", weight: 10},
		"фольг":            {name: "Товары для дома", weight: 10},
		"пленк":            {name: "Товары для дома", weight: 10},
		"губк":             {name: "Товары для дома", weight: 10},
		"тряпк":            {name: "Товары для дома", weight: 10},
		"салфетк":          {name: "Товары для дома", weight: 10},
		"ламп":             {name: "Товары для дома", weight: 10},
		"батарейк":         {name: "Товары для дома", weight: 10},
		"свеч":             {name: "Товары для дома", weight: 10},
		"ведр":             {name: "Товары для дома", weight: 10},
		"швабр":            {name: "Товары для дома", weight: 10},
		"коврик":           {name: "Товары для дома", weight: 10},
		"плечик":           {name: "Товары для дома", weight: 10},
		"органайзер":       {name: "Товары для дома", weight: 10},
		"полотенц":         {name: "Товары для дома", weight: 10},
		"постельн":         {name: "Товары для дома", weight: 10},
		"подушк":           {name: "Товары для дома", weight: 10},
		"одеял":            {name: "Товары для дома", weight: 10},
		"порошок":          {name: "Бытовая химия", weight: 10},
		"стирк":            {name: "Бытовая химия", weight: 10},
		"кондиционер бель": {name: "Бытовая химия", weight: 10},
		"отбелив":          {name: "Бытовая химия", weight: 10},
		"пятновывод":       {name: "Бытовая химия", weight: 10},
		"средство":         {name: "Бытовая химия", weight: 10},
		"чист":             {name: "Бытовая химия", weight: 10},
		"моющ":             {name: "Бытовая химия", weight: 10},
		"посуд":            {name: "Бытовая химия", weight: 10},
		"унитаз":           {name: "Бытовая химия", weight: 10},
		"ванн":             {name: "Бытовая химия", weight: 10},
		"плит":             {name: "Бытовая химия", weight: 10},
		"жир":              {name: "Бытовая химия", weight: 10},
		"стекл":            {name: "Бытовая химия", weight: 10},
		"освежитель":       {name: "Бытовая химия", weight: 10},
		"дезинфиц":         {name: "Бытовая химия", weight: 10},
		"доместос":         {name: "Бытовая химия", weight: 10},
		"fairy":            {name: "Бытовая химия", weight: 10},
		"фейри":            {name: "Бытовая химия", weight: 10},
		"персил":           {name: "Бытовая химия", weight: 10},
		"ариэль":           {name: "Бытовая химия", weight: 10},
		"tide":             {name: "Бытовая химия", weight: 10},
		"тайд":             {name: "Бытовая химия", weight: 10},
		"футболк":          {name: "Одежда и обувь", weight: 10},
		"рубашк":           {name: "Одежда и обувь", weight: 10},
		"брюк":             {name: "Одежда и обувь", weight: 10},
		"джинс":            {name: "Одежда и обувь", weight: 10},
		"шорт":             {name: "Одежда и обувь", weight: 10},
		"плать":            {name: "Одежда и обувь", weight: 10},
		"юбк":              {name: "Одежда и обувь", weight: 10},
		"куртк":            {name: "Одежда и обувь", weight: 10},
		"пальто":           {name: "Одежда и обувь", weight: 10},
		"носок":            {name: "Одежда и обувь", weight: 10},
		"трус":             {name: "Одежда и обувь", weight: 10},
		"бель":             {name: "Одежда и обувь", weight: 10},
		"колгот":           {name: "Одежда и обувь", weight: 10},
		"кроссовк":         {name: "Одежда и обувь", weight: 10},
		"ботинк":           {name: "Одежда и обувь", weight: 10},
		"туфл":             {name: "Одежда и обувь", weight: 10},
		"сапог":            {name: "Одежда и обувь", weight: 10},
		"тапоч":            {name: "Одежда и обувь", weight: 10},
		"шнурок":           {name: "Одежда и обувь", weight: 10},
		"ремень":           {name: "Одежда и обувь", weight: 10},
		"шапк":             {name: "Одежда и обувь", weight: 10},
		"перчатк":          {name: "Одежда и обувь", weight: 10},
		"телефон":          {name: "Техника и электроника", weight: 10},
		"смартфон":         {name: "Техника и электроника", weight: 10},
		"ноутбук":          {name: "Техника и электроника", weight: 10},
		"планшет":          {name: "Техника и электроника", weight: 10},
		"заряд":            {name: "Техника и электроника", weight: 10},
		"кабель":           {name: "Техника и электроника", weight: 10},
		"usb":              {name: "Техника и электроника", weight: 10},
		"type-c":           {name: "Техника и электроника", weight: 10},
		"наушник":          {name: "Техника и электроника", weight: 10},
		"колонк":           {name: "Техника и электроника", weight: 10},
		"мыш":              {name: "Техника и электроника", weight: 10},
		"клавиатур":        {name: "Техника и электроника", weight: 10},
		"монитор":          {name: "Техника и электроника", weight: 10},
		"адаптер":          {name: "Техника и электроника", weight: 10},
		"переходник":       {name: "Техника и электроника", weight: 10},
		"флеш":             {name: "Техника и электроника", weight: 10},
		"карта памяти":     {name: "Техника и электроника", weight: 10},
		"роутер":           {name: "Техника и электроника", weight: 10},
		"блендер":          {name: "Техника и электроника", weight: 10},
		"чайник":           {name: "Техника и электроника", weight: 10},
		"утюг":             {name: "Техника и электроника", weight: 10},
		"фен":              {name: "Техника и электроника", weight: 10},
		"пылесос":          {name: "Техника и электроника", weight: 10},
		"миксер":           {name: "Техника и электроника", weight: 10},
		"лекарств":         {name: "Здоровье", weight: 10},
		"таблет":           {name: "Здоровье", weight: 10},
		"капсул":           {name: "Здоровье", weight: 10},
		"сироп":            {name: "Здоровье", weight: 10},
		"мазь":             {name: "Здоровье", weight: 10},
		"спрей":            {name: "Здоровье", weight: 10},
		"капл":             {name: "Здоровье", weight: 10},
		"витамин":          {name: "Здоровье", weight: 10},
		"бинт":             {name: "Здоровье", weight: 10},
		"пластыр":          {name: "Здоровье", weight: 10},
		"термометр":        {name: "Здоровье", weight: 10},
		"маск медицин":     {name: "Здоровье", weight: 10},
		"антисептик":       {name: "Здоровье", weight: 10},
		"ибупрофен":        {name: "Здоровье", weight: 10},
		"парацетамол":      {name: "Здоровье", weight: 10},
		"цитрамон":         {name: "Здоровье", weight: 10},
		"аспирин":          {name: "Здоровье", weight: 10},
		"ношпа":            {name: "Здоровье", weight: 10},
		"активирован":      {name: "Здоровье", weight: 10},
		"уголь":            {name: "Здоровье", weight: 10},
		"кино":             {name: "Развлечения", weight: 10},
		"билет":            {name: "Развлечения", weight: 10},
		"театр":            {name: "Развлечения", weight: 10},
		"концерт":          {name: "Развлечения", weight: 10},
		"музе":             {name: "Развлечения", weight: 10},
		"игр":              {name: "Развлечения", weight: 10},
		"настольн":         {name: "Развлечения", weight: 10},
		"книг":             {name: "Развлечения", weight: 10},
		"журнал":           {name: "Развлечения", weight: 10},
		"подписк":          {name: "Развлечения", weight: 10},
		"netflix":          {name: "Развлечения", weight: 10},
		"spotify":          {name: "Развлечения", weight: 10},
		"steam":            {name: "Развлечения", weight: 10},
		"playstation":      {name: "Развлечения", weight: 10},
		"xbox":             {name: "Развлечения", weight: 10},
		"кафе":             {name: "Развлечения", weight: 10},
		"ресторан":         {name: "Развлечения", weight: 10},
		"бар":              {name: "Развлечения", weight: 10},
		"доставка":         {name: "Развлечения", weight: 10},
		"пицц":             {name: "Развлечения", weight: 10},
		"суши":             {name: "Развлечения", weight: 10},
	}

	entries := make([]dictionaryEntry, 0, len(raw))
	for keyword, entry := range raw {
		entries = append(entries, dictionaryEntry{keyword: keyword, entry: entry})
	}

	slices.SortFunc(entries, func(a, b dictionaryEntry) int {
		return cmp.Compare(utf8.RuneCountInString(b.keyword), utf8.RuneCountInString(a.keyword))
	})

	return entries
}
