package calculator

import (
	"math"
	"strconv"

	"github.com/KevinMarklin/decision-pause/backend/internal/model"
)

type marketingCase struct {
	key, title, emoji string
	boostFactor, costFactor float64
}

var marketingCases = []marketingCase{
	{key: "expected", title: "Ожидаемый", emoji: "🟢", boostFactor: 1, costFactor: 1},
	{key: "moderate", title: "Умеренно негативный", emoji: "🟡", boostFactor: 0.85, costFactor: 1.15},
	{key: "negative", title: "Негативный", emoji: "🔴", boostFactor: 0.50, costFactor: 1.35},
}

func MarketingResult(in model.MarketingInput) model.MarketingResult {
	margin := in.AvgCheck * in.MarginPct / 100
	breakEvenOrders := 0
	if margin > 0 {
		breakEvenOrders = int(math.Ceil(in.Budget / margin))
	}

	result := model.MarketingResult{
		BreakEvenOrders:  breakEvenOrders,
		BreakEvenRevenue: float64(breakEvenOrders) * in.AvgCheck,
		Scenarios:        make([]model.MarketingScenario, 0, len(marketingCases)),
		Consequences:     []string{},
		Checklist: []string{
			"Проверил стоимость лида на небольшом тестовом бюджете",
			"Понимаю, сколько заказов нужно для окупаемости",
			"Заложил резерв на просадку конверсии и рост CAC",
			"Определил лимит расходов и точку остановки кампании",
			"Есть план обработки новых лидов и заказов",
		},
	}

	for _, c := range marketingCases {
		boost := in.ExpectedBoostPct * c.boostFactor
		leadCost := in.LeadCost * c.costFactor
		orders := 0.0
		if leadCost > 0 {
			orders = in.Budget / leadCost * (1 + boost/100)
		}
		adRevenue := orders * in.AvgCheck
		adProfit := orders*margin - in.Budget
		cashFlow := in.CurrentRevenue + adRevenue - in.FixedExpenses - in.Budget
		status := "green"
		if cashFlow < 0 {
			status = "red"
		} else if adProfit < 0 {
			status = "yellow"
		}
		var reserveMonths *int
		if cashFlow < 0 {
			months := int(math.Floor(in.Reserve / math.Abs(cashFlow)))
			reserveMonths = &months
		}
		result.Scenarios = append(result.Scenarios, model.MarketingScenario{
			Key: c.key, Title: c.title, Emoji: c.emoji, AdditionalOrders: orders,
			AdRevenue: adRevenue, AdProfit: adProfit, ROI: adProfit / in.Budget * 100,
			CashFlow: cashFlow, Status: status, ReserveMonths: reserveMonths,
			Consequence: marketingConsequence(status, cashFlow, reserveMonths),
		})
	}
	result.Consequences = marketingConsequences(result.Scenarios)
	return result
}

func marketingConsequence(status string, cashFlow float64, months *int) string {
	if status == "red" {
		return "Кассовый разрыв: при таком потоке резерв покрывает примерно " + formatIntPtr(months) + " мес."
	}
	if status == "yellow" {
		return "Реклама работает в минус, но основной бизнес пока покрывает дефицит."
	}
	return "Реклама добавляет положительный поток, кассовый разрыв не формируется."
}

func formatIntPtr(v *int) string {
	if v == nil { return "0" }
	return formatInt(*v)
}

func formatInt(v int) string {
	if v == 1 { return "1" }
	return fmtInt(v)
}

func fmtInt(v int) string {
	return strconv.Itoa(v)
}

func marketingConsequences(scenarios []model.MarketingScenario) []string {
	result := make([]string, 0, len(scenarios))
	for _, scenario := range scenarios { result = append(result, scenario.Consequence) }
	return result
}
