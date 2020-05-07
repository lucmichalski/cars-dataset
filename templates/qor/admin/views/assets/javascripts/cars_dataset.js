/*! waitForImages jQuery Plugin 2018-02-13 */
!function(a){"function"==typeof define&&define.amd?define(["jquery"],a):"object"==typeof exports?module.exports=a(require("jquery")):a(jQuery)}(function(a){var b="waitForImages",c=function(a){return a.srcset&&a.sizes}(new Image);a.waitForImages={hasImageProperties:["backgroundImage","listStyleImage","borderImage","borderCornerImage","cursor"],hasImageAttributes:["srcset"]},a.expr.pseudos["has-src"]=function(b){return a(b).is('img[src][src!=""]')},a.expr.pseudos.uncached=function(b){return!!a(b).is(":has-src")&&!b.complete},a.fn.waitForImages=function(){var d,e,f,g=0,h=0,i=a.Deferred(),j=this,k=[],l=a.waitForImages.hasImageProperties||[],m=a.waitForImages.hasImageAttributes||[],n=/url\(\s*(['"]?)(.*?)\1\s*\)/g;if(a.isPlainObject(arguments[0])?(f=arguments[0].waitForAll,e=arguments[0].each,d=arguments[0].finished):1===arguments.length&&"boolean"===a.type(arguments[0])?f=arguments[0]:(d=arguments[0],e=arguments[1],f=arguments[2]),d=d||a.noop,e=e||a.noop,f=!!f,!a.isFunction(d)||!a.isFunction(e))throw new TypeError("An invalid callback was supplied.");return this.each(function(){var b=a(this);f?b.find("*").addBack().each(function(){var b=a(this);b.is("img:has-src")&&!b.is("[srcset]")&&k.push({src:b.attr("src"),element:b[0]}),a.each(l,function(a,c){var d,e=b.css(c);if(!e)return!0;for(;d=n.exec(e);)k.push({src:d[2],element:b[0]})}),a.each(m,function(a,c){var d=b.attr(c);return!d||void k.push({src:b.attr("src"),srcset:b.attr("srcset"),element:b[0]})})}):b.find("img:has-src").each(function(){k.push({src:this.src,element:this})})}),g=k.length,h=0,0===g&&(d.call(j),i.resolveWith(j)),a.each(k,function(f,k){var l=new Image,m="load."+b+" error."+b;a(l).one(m,function b(c){var f=[h,g,"load"==c.type];if(h++,e.apply(k.element,f),i.notifyWith(k.element,f),a(this).off(m,b),h==g)return d.call(j[0]),i.resolveWith(j[0]),!1}),c&&k.srcset&&(l.srcset=k.srcset,l.sizes=k.sizes),l.src=k.src}),i.promise()}});

var VehiclesChart, VehicleImagesChart;

function RenderChart(vehiclesData, vehiclesImagesData) {
    Chart.defaults.global.responsive = true;

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

    // Vehicles Images 
    var vehicleImagesDateLables = [];
    var vehicleImagesCounts = [];
    for (var i = 0; i < vehiclesImagesData.length; i++) {
        vehicleImagesDateLables.push(vehiclesImagesData[i].Date.substring(5,10));
        vehicleImagesCounts.push(vehiclesImagesData[i].Total)
    }
    if(VehicleImagesChart){
        VehicleImagesChart.destroy();
    }
    var vehicle_images_context = document.getElementById("vehicle_images_report").getContext("2d");
    var vehicle_images_data = ChartData(vehicleImagesDateLables, vehicleImagesCounts);
    VehicleImagesChart = new Chart(vehicle_images_context).Line(vehicle_images_data, "");

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
      RenderChart(jsonData.Vehicles, jsonData.VehicleImages);
      $("#vehicles_report_loader").hide();
      $("#vehicle_images_report_loader").hide();
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

  var gridImages = $("p[data-heading=File] img")
  console.log("gridImages:", gridImages);

  var generateBoundingBoxes = function(str) {
    var gridImages = $("p[data-heading=File] img")
    for (let i = 0; i < gridImages.length; i++) {
      gridImages[i].src = 'http://51.91.21.67:9008'+ gridImages[i].src
    }    
  };

  // generateBoundingBoxes()

  /*
    // ref.
    // https://stackoverflow.com/questions/4839993/how-to-draw-polygons-on-an-html5-canvas
    // https://www.w3schools.com/tags/canvas_rect.asp
    var c = document.getElementById("myCanvas");
    var ctx = c.getContext("2d");
    ctx.beginPath();
    ctx.rect(20, 20, 150, 100);
    ctx.stroke();
  */

});

var getRealWidth = function(src) {
    var dim = {}
    $("<img>").attr("src", src).on("load", function() {
        realWidth = this.width;
        dim.realWidth = realWidth
        realHeight = this.height;
        dim.realHeight = realHeight
        console.log("Original width=" + realWidth + ", " + "Original height=" + realHeight);
    });
    return dim
};

var generateBoundingBoxes = function() {

  $(".mdl-card__supporting-text.qor-table--ml-slideout").each(function( index ) {
      var bboxElement = $(this).find('p[data-heading=BBox]').text()
      var bbox = bboxElement.split(",");
      console.log(bbox)
      var topX = bbox[0]
      var topY = bbox[1]
      var bottomX = bbox[2]
      var bottomY = bbox[3]
      var myImg = $(this).find('img')
      console.log("myImg:", myImg[0].width, myImg[0].width)
      console.log("myImg.width:", $(myImg).width())
      console.log("myImg.height:", $(myImg).height())
      console.log("src=", $(myImg).attr("src"))
      var realWidth = "";
      var realHeight = "";
      dim = getRealWidth($(myImg).attr("src"))

      console.log("bbox", bbox)
      console.log("realWidth", dim.realWidth, "realHeight", dim.realHeight)

      widthRatio = $(myImg).width() * 100 / dim.realWidth;
      heightRatio = $(myImg).height() * 100 / dim.realHeight;

      console.log("widthRatio=", widthRatio)
      console.log("heightRatio=", heightRatio)

      var canvas = document.createElement("canvas")
      var ctx = canvas.getContext('2d');
      // ctx.fillStyle = '#f00';
      ctx.beginPath();
      ctx.rect(topX, topY, topX-bottomX, topY-bottomY);
      ctx.strokeStyle = "#5eba7d";
      ctx.lineWidth = 1;
      ctx.stroke();
      $(this).append(canvas)
  });
};

$(window).waitForImages(function() {
// $(window).waitForImages().done(function() {
  generateBoundingBoxes();
});

  // $("canvas[id=cvs-1222517] p[data-heading=BBox]")

  /*
    // ref.
    // https://stackoverflow.com/questions/4839993/how-to-draw-polygons-on-an-html5-canvas
    // https://www.w3schools.com/tags/canvas_rect.asp
    var c = document.getElementById("myCanvas");
    var ctx = c.getContext("2d");
    ctx.beginPath();
    ctx.rect(20, 20, 150, 100);
    ctx.stroke();
  */
