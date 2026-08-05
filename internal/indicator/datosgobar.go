package indicator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

// seriesAPIBase is the official, no-auth "Series de Tiempo" API run by the
// Argentine government (datos.gob.ar), documented at
// https://datosgobar.github.io/series-tiempo-ar-api/. It re-publishes
// INDEC's IPC series among ~30k other official indicators.
const seriesAPIBase = "https://apis.datos.gob.ar/series/api/series/"

// DefaultIPCSeriesID is INDEC's national headline CPI ("IPC Nivel General
// Nacional", base December 2016 = 100), published by the Secretaría de
// Programación Macroeconómica under dataset 145, distribution 145.3.
// Verified directly against the catalog metadata
// (infra.datos.gob.ar/catalog/sspm/data.json, dataset 145 → distribution
// 145.3 → field 145.3_INGNACNAL_DICI_M_15) and against the live series API,
// not guessed — config.Load defaults INDICATOR_IPC_SERIES_ID to this value.
const DefaultIPCSeriesID = "145.3_INGNACNAL_DICI_M_15"

// seriesAPILimit must be large enough to cover the whole monthly series
// since Dec 2016; the API's default page size is only 100 points, which
// silently truncates to the *oldest* 100 months and drops everything recent.
const seriesAPILimit = 5000

// Each data point is a heterogeneous tuple: ["2016-12-01", 100.0] — a JSON
// string period followed by a JSON number value, not two numbers.
type seriesAPIResponse struct {
	Data [][2]json.RawMessage `json:"data"`
}

// SyncIPC fetches the configured IPC series from datos.gob.ar and stores
// every monthly value. UpsertIndicator is keyed on (code, period), so
// re-running this is always safe and just refreshes existing months.
func (s *Service) SyncIPC(ctx context.Context, seriesID string) (int, error) {
	if seriesID == "" {
		return 0, fmt.Errorf("no IPC series id configured: set INDICATOR_IPC_SERIES_ID (find it via https://apis.datos.gob.ar/series/api/search/?q=ipc+nivel+general)")
	}

	url := fmt.Sprintf("%s?ids=%s&format=json&limit=%d", seriesAPIBase, seriesID, seriesAPILimit)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("datos.gob.ar returned status %d", resp.StatusCode)
	}

	var parsed seriesAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return 0, fmt.Errorf("decoding datos.gob.ar response: %w", err)
	}

	count := 0
	for _, point := range parsed.Data {
		var periodStr string
		if err := json.Unmarshal(point[0], &periodStr); err != nil {
			continue
		}
		period, err := time.Parse("2006-01-02", periodStr)
		if err != nil {
			continue
		}

		var value decimal.Decimal
		if err := json.Unmarshal(point[1], &value); err != nil {
			continue
		}

		if _, err := s.upsert(ctx, IPCCode, period, value, "datos.gob.ar"); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
