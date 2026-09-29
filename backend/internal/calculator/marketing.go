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
		Budget:           in.Budget,
		Reserve:          in.Reserve,
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
		adCost := in.Budget
		adProfit := orders*margin - in.Budget
		revenue := in.CurrentRevenue + adRevenue
		expenses := in.FixedExpenses + adCost
		netProfit := revenue - expenses
		cashFlow := netProfit
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
			AdRevenue: adRevenue, AdCost: adCost, AdProfit: adProfit, ROI: safePct(adProfit, adCost),
			CashFlow: cashFlow, Status: status, ReserveMonths: reserveMonths,
			Consequence: marketingConsequence(status, cashFlow, reserveMonths),
			Revenue: revenue, Expenses: expenses, NetProfit: netProfit,
			MarginPct: safePct(netProfit, revenue),
		})
	}
	result.Consequences = marketingConsequences(result.Scenarios)
	if in.IsBlackSwan {
		return ApplyBlackSwan(result)
	}
	return result
}

// ApplyBlackSwan applies a deterministic crisis shock to an already calculated result.
func ApplyBlackSwan(res model.MarketingResult) model.MarketingResult {
	res.IsBlackSwanActive = true
	for i := range res.Scenarios {
		s := &res.Scenarios[i]
		s.Revenue *= 0.80
		s.AdRevenue *= 0.80
		s.AdditionalOrders *= 0.80
		s.AdCost = res.Budget * 1.15
		s.Expenses *= 1.15
		s.AdProfit = (s.AdProfit+s.AdCost/1.15)*0.80 - s.AdCost
		s.NetProfit = s.Revenue - s.Expenses
		s.CashFlow = s.NetProfit
		s.MarginPct = safePct(s.NetProfit, s.Revenue)
		s.ROI = safePct(s.AdProfit, s.AdCost)
		if s.CashFlow < 0 {
			months := 0
			if s.CashFlow != 0 {
				months = int(math.Floor(res.Reserve / math.Abs(s.CashFlow)))
			}
			s.ReserveMonths = &months
			s.Status = "red"
		} else if s.AdProfit < 0 {
			s.Status = "yellow"
		} else {
			s.Status = "green"
		}
		s.Consequence = marketingConsequence(s.Status, s.CashFlow, s.ReserveMonths)
	}
	res.Consequences = marketingConsequences(res.Scenarios)
	return res
}

func safePct(value, base float64) float64 {
	if base == 0 {
		return 0
	}
	return value / base * 100
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
