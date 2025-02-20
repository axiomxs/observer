package trace

import (
	"net/http"

	"github.com/bclswl0827/observer/app"
	"github.com/bclswl0827/observer/server/response"
	"github.com/gin-gonic/gin"
)

// @Summary Observer event trace
// @Description Get list of earthquake events data source and earthquake events from specified data source
// @Router /trace [post]
// @Accept application/x-www-form-urlencoded
// @Produce application/json
// @Param source formData string true "Use `show` to get available sources first, then choose one and request again to get events"
// @Failure 400 {object} response.HttpResponse "Failed to read earthquake event list due to invalid data source"
// @Failure 500 {object} response.HttpResponse "Failed to read earthquake event list due to failed to read data source"
// @Success 200 {object} response.HttpResponse{data=[]Event} "Successfully read the list of earthquake events"
func (t *Trace) RegisterModule(rg *gin.RouterGroup, options *app.ServerOptions) {
	sources := []DataSource{
		&SCEA_E{}, &SCEA_B{}, &CEIC{},
		&USGS{}, &JMA{}, &CWA{}, &HKO{},
	}

	rg.POST("/trace", func(c *gin.Context) {
		var binding Binding
		if err := c.ShouldBind(&binding); err != nil {
			response.Error(c, http.StatusBadRequest)
			return
		}

		if binding.Source == "show" {
			type availableSources struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			}

			var list []availableSources
			for _, v := range sources {
				name, value := v.Property()
				list = append(list, availableSources{
					Name:  name,
					Value: value,
				})
			}

			response.Message(c, "Successfully read available data source list", list)
			return
		}

		var (
			latitude  = options.FeatureOptions.Config.Station.Latitude
			longitude = options.FeatureOptions.Config.Station.Longitude
		)
		for _, v := range sources {
			_, value := v.Property()
			if value == binding.Source {
				events, err := v.List(latitude, longitude)
				if err != nil {
					response.Error(c, http.StatusInternalServerError)
					return
				}

				sortByTimestamp(events)
				response.Message(c, "Successfully read the list of earthquake events", events)
				return
			}
		}

		response.Error(c, http.StatusBadRequest)
	})
}
