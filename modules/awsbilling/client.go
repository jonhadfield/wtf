package awsbilling

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/costexplorer"
)

func getSession(acc account) (*session.Session, error) {
	creds := credentials.NewStaticCredentials(acc.accessKeyId,
		acc.accessKeySecret, "")
	return session.NewSession(&aws.Config{Credentials: creds})
}

func getCostForPeriod(ce *costexplorer.CostExplorer, startDate, endDate string) (cost float64, err error) {
	di := costexplorer.DateInterval{
		End:   &startDate,
		Start: &endDate,
	}

	var ceo *costexplorer.GetCostAndUsageOutput
	ceo, err = ce.GetCostAndUsage(&costexplorer.GetCostAndUsageInput{
		Granularity: strPtr("MONTHLY"),
		Metrics:     []*string{strPtr("UnblendedCost")},
		TimePeriod:  &di,
	})

	if err != nil {
		return
	}

	rbt := ceo.ResultsByTime
	mtdAmount := rbt[0].Total["UnblendedCost"].Amount

	cost, err = strconv.ParseFloat(*mtdAmount, 64)

	return
}

func getForecastedCost(ce *costexplorer.CostExplorer, lastDOM string) (forecast float64, err error) {
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	fi := costexplorer.DateInterval{
		End:   &lastDOM,
		Start: &tomorrow,
	}

	pil := int64(75)
	gfi := costexplorer.GetCostForecastInput{
		Granularity:             strPtr("MONTHLY"),
		Metric:                  strPtr("AMORTIZED_COST"),
		PredictionIntervalLevel: &pil,
		TimePeriod:              &fi,
	}

	var gfo *costexplorer.GetCostForecastOutput

	gfo, err = ce.GetCostForecast(&gfi)
	if err != nil {
		return
	}

	var lowS, highS string
	lowS = *gfo.ForecastResultsByTime[0].PredictionIntervalLowerBound
	highS = *gfo.ForecastResultsByTime[0].PredictionIntervalUpperBound

	var low, high float64

	low, err = strconv.ParseFloat(lowS, 64)
	if err != nil {
		return
	}

	high, err = strconv.ParseFloat(highS, 64)

	forecast = (low + high) / 2

	return forecast, err
}

func getBillingDetail(acc account) (current, lastMonth, forecastAvg float64, err error) {
	var sess *session.Session

	sess, err = getSession(acc)
	if err != nil {
		return
	}

	ce := costexplorer.New(sess)

	y, m, _ := time.Now().Date()
	firstDOMt := time.Date(y, m, 1, 0, 0, 0, 0, time.Now().Location())
	firstDOMs := firstDOMt.Format("2006-01-02")
	lastDOMt := firstDOMt.AddDate(0, 1, 0).Add(-time.Nanosecond)
	lastDOMs := lastDOMt.Format("2006-01-02")

	current, err = getCostForPeriod(ce, lastDOMs, firstDOMs)
	if err != nil {
		return
	}

	var forecast float64

	forecast, err = getForecastedCost(ce, lastDOMs)
	if err != nil {
		return
	}

	y, m, _ = time.Now().AddDate(0, -1, 0).Date()
	firstDOMt = time.Date(y, m, 1, 0, 0, 0, 0, time.Now().Location())
	firstDOMs = firstDOMt.Format("2006-01-02")
	lastDOMt = firstDOMt.AddDate(0, 1, 0).Add(-time.Nanosecond)
	lastDOMs = lastDOMt.Format("2006-01-02")

	lastMonth, err = getCostForPeriod(ce, lastDOMs, firstDOMs)
	if err != nil {
		return
	}

	return current, lastMonth, forecast, err
}

func getAccountOutput(widget *Widget, acc account) (out string, err error) {
	var current, lastMonth, forecastAvg float64

	current, lastMonth, forecastAvg, err = getBillingDetail(acc)
	if err != nil {
		return
	}

	alias := trim(acc.alias, widget.settings.aliasWidth)

	var colourPrefix string

	switch {
	case current > acc.budget:
		colourPrefix = "[red]"
	case current < acc.budget && forecastAvg > acc.budget:
		colourPrefix = "[yellow]"
	default:
		colourPrefix = "[green]"
	}

	switch widget.settings.output {
	case "detailed":
		//out = getDetailedOutput(current, forecastAvg, lastMonth, acc, colourPrefix, alias)
		switch {
		case current < 1000.00 && acc.budget < 1000.00 && forecastAvg < 1000.00:
			out = fmt.Sprintf("$%-6.2f budget $%-6.2f forecast $%-6.2f last month $%-6.2f",
				current, acc.budget, forecastAvg, lastMonth)
		case current < 10000.00 && acc.budget < 10000.00 && forecastAvg < 10000.00:
			out = fmt.Sprintf("$%-7.2f budget $%-7.2f forecast $%-7.2f last month $%-7.2f",
				current, acc.budget, forecastAvg, lastMonth)
		case current < 100000.00 && acc.budget < 100000.00 && forecastAvg < 100000.00:
			out = fmt.Sprintf("$%-8.2f budget $%-8.2f forecast $%-8.2f last month $%-8.2f",
				current, acc.budget, forecastAvg, current)
		default:
			out = fmt.Sprintf("$%-9.2f budget $%-9.2f forecast $%-9.2f last month $%-9.2f",
				current, acc.budget, forecastAvg, lastMonth)
		}
	case "minimal":
		switch {
		case current > acc.budget:
			out = fmt.Sprintf("$%.2f", current)
		case current < acc.budget && forecastAvg > acc.budget:
			out = fmt.Sprintf("$%.2f",
				current)
		default:
			out = fmt.Sprintf("$%.2f", current)
		}
	case "default":
		switch {
		case current > acc.budget:
			out = fmt.Sprintf("$%.2f | budget $%.2f", current, acc.budget)
		case current < acc.budget && forecastAvg > acc.budget:
			out = fmt.Sprintf("$%.2f | forecast $%.2f", current, forecastAvg)
		default:
			out = fmt.Sprintf("$%.2f", current)
		}
	}

	// add account name prefix
	out = fmt.Sprintf("%s %s %s", colourPrefix, alias, out)

	return out, err
}

type account struct {
	alias           string
	accessKeyId     string
	accessKeySecret string
	budget          float64
}

func trim(i string, width int) string {
	if len(i) > width {
		return i[:width] + ".."
	}

	return i
}
