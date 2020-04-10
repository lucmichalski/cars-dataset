var ProductsChart,ReviewsChart,VehiclesChart;
function RenderChart(productsData, reviewsData, vehiclesData) {
    Chart.defaults.global.responsive = true;

    // Products
    var productDateLables = [];
    var productCounts = [];
    for (var i = 0; i < productsData.length; i++) {
        productDateLables.push(productsData[i].Date.substring(5,10));
        productCounts.push(productsData[i].Total)
    }
    if(ProductsChart){
        ProductsChart.destroy();
    }
    var products_context = document.getElementById("products_report").getContext("2d");
    var products_data = ChartData(productDateLables,productCounts);
    ProductsChart = new Chart(products_context).Line(products_data, "");

    // Vehicles
    var vehicleDateLables = [];
    var vehicleCounts = [];
    for (var i = 0; i < vehiclesData.length; i++) {
        vehicleDateLables.push(vehiclesData[i].Date.substring(5,10));
        vehicleCounts.push(vehiclesData[i].Total)
    }
    if(VehiclesChart){
        VehiclesChart.destroy();
    }
    var vehicles_context = document.getElementById("vehicles_report").getContext("2d");
    var vehicles_data = ChartData(vehicleDateLables,vehicleCounts);
    VehiclesChart = new Chart(vehicles_context).Line(vehicles_data, "");

    // Reviews
    var reviewsDateLables = [];
    var reviewsCounts = [];
    for (var i = 0; i < reviewsData.length; i++) {
        reviewsDateLables.push(reviewsData[i].Date.substring(5,10));
        reviewsCounts.push(reviewsData[i].Total)
    }
    if(ReviewsChart){
        ReviewsChart.destroy();
    }
    var reviews_context = document.getElementById("reviews_report").getContext("2d");
    var reviews_data = ChartData(reviewsDateLables,reviewsCounts);
    ReviewsChart = new Chart(reviews_context).Bar(reviews_data, "");

}

function ChartData(lables, counts) {
    var chartData = {
      labels: lables,
      datasets: [
      {
        label: "Users Report",
        fillColor: "rgba(151,187,205,0.2)",
        strokeColor: "rgba(151,187,205,1)",
        pointColor: "rgba(151,187,205,1)",
        pointStrokeColor: "#fff",
        pointHighlightFill: "#fff",
        pointHighlightStroke: "rgba(151,187,205,1)",
        data: counts
      }
      ]
    };
    return chartData;
}

Date.prototype.Format = function (fmt) {
    var o = {
        "M+": this.getMonth() + 1,
        "d+": this.getDate(),
        "h+": this.getHours(),
        "m+": this.getMinutes(),
        "s+": this.getSeconds(),
        "q+": Math.floor((this.getMonth() + 3) / 3),
        "S": this.getMilliseconds()
    };
    if (/(y+)/.test(fmt)) fmt = fmt.replace(RegExp.$1, (this.getFullYear() + "").substr(4 - RegExp.$1.length));
    for (var k in o)
    if (new RegExp("(" + k + ")").test(fmt)) fmt = fmt.replace(RegExp.$1, (RegExp.$1.length == 1) ? (o[k]) : (("00" + o[k]).substr(("" + o[k]).length)));
    return fmt;
}

Date.prototype.AddDate = function (add){
    var date = this.valueOf();
    date = date + add * 24 * 60 * 60 * 1000
    date = new Date(date)
    return date;
}

// qor dashboard
$(document).ready(function() {
  var yesterday = (new Date()).AddDate(-1);
  var defStartDate = yesterday.AddDate(-6);
  $("#startDate").val(defStartDate.Format("yyyy-MM-dd"));
  $("#endDate").val(yesterday.Format("yyyy-MM-dd"));
  $(".j-update-record").click(function(){
    $.getJSON("/admin/reports.json",{startDate:$("#startDate").val(), endDate:$("#endDate").val()},function(jsonData){
      RenderChart(jsonData.Products,jsonData.Reviews,jsonData.Vehicles);
      $("#products_report_loader").hide();
      $("#reviews_report_loader").hide();
      $("#vehicles_report_loader").hide();
    });
  });
  $(".j-update-record").click();

  $(".yesterday-reports").click(function() {
    $("#startDate").val(yesterday.Format("yyyy-MM-dd"));
    $("#endDate").val(yesterday.Format("yyyy-MM-dd"));
    $(".j-update-record").click();
    $(this).blur();
  });

  $(".this-week-reports").click(function() {
    var beginningOfThisWeek = yesterday.AddDate(-yesterday.getDay() + 1)
    $("#startDate").val(beginningOfThisWeek.Format("yyyy-MM-dd"));
    $("#endDate").val(beginningOfThisWeek.AddDate(6).Format("yyyy-MM-dd"));
    $(".j-update-record").click();
    $(this).blur();
  });

  $(".last-week-reports").click(function() {
    var endOfLastWeek = yesterday.AddDate(-yesterday.getDay())
    $("#startDate").val(endOfLastWeek.AddDate(-6).Format("yyyy-MM-dd"));
    $("#endDate").val(endOfLastWeek.Format("yyyy-MM-dd"));
    $(".j-update-record").click();
    $(this).blur();
  });
});
