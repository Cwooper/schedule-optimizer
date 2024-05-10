const WIDTH = 1000;
const HEIGHT = 600;
const TOP_MARGIN = 0.07;
const SIDE_MARGIN = 0.1;
const MAJOR_LINE_WIDTH = 6;
const MINOR_LINE_WIDTH = 2;
const DAY_COUNT = 5;

var weekDayStrings = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"];
var weekDayChars = ["m", "t", "w", "r", "f", "s", "u"];

var courseCount = 1;
                
var startTime = 0;
var endTime = 0; 
var minCredits = 0;
var maxCredits = 0;
var minCourses = 0;
var maxCourses = 0;
var mandatoryGroups = [];

var currentOptionViewed = undefined;
var allSchedules = undefined;
    
    
    //TODO: Add information about the schedule in the top 
    function drawSchedule(scheduleIndex){        
        var canvas = document.getElementById("MainCanvas");
        var ctx = canvas.getContext("2d");

        canvas.width = WIDTH;
        canvas.height = HEIGHT;

        var dayWidth = (WIDTH*(1-SIDE_MARGIN))/DAY_COUNT;
        
        var hourCount = (endTime - startTime)/60.0;
        
        var hourHeight = (HEIGHT*(1-TOP_MARGIN)) / hourCount;

        var topMarginOffset = HEIGHT*TOP_MARGIN
        var sideMarginOffset = WIDTH*SIDE_MARGIN


        let schedule = allSchedules[scheduleIndex];

        
        if(schedule === undefined){


            var fontSize = WIDTH/30;
            ctx.font = fontSize + "px Arial";

            
            ctx.fillText("No options found.", WIDTH/3, HEIGHT/3);


            ctx.fillText("Things to try: ", WIDTH/4, HEIGHT/3 + fontSize*2);
            ctx.fillText("  • Change course count or credit requirements.", WIDTH/4, HEIGHT/3 + fontSize*3);
            ctx.fillText("  • Remove mandatory groups.", WIDTH/4, HEIGHT/3 + fontSize*4);
            ctx.fillText("  • Add more sections.", WIDTH/4, HEIGHT/3 + fontSize*5);

        }
        else{

            ctx.fillStyle = "#FFFFFF";
            ctx.fillRect(0, 0, WIDTH, HEIGHT);
            
            ctx.lineWidth = MAJOR_LINE_WIDTH;
            ctx.fillStyle = "#000000";
            
            //Create top margin
            ctx.beginPath();
            ctx.moveTo(sideMarginOffset , 0);
            ctx.lineTo(sideMarginOffset, HEIGHT);
            ctx.stroke();

            ctx.beginPath();
            ctx.moveTo(0 , topMarginOffset);
            ctx.lineTo(WIDTH, topMarginOffset);
            ctx.stroke();

            ctx.beginPath();
            ctx.moveTo(0 , 0);
            ctx.lineTo(WIDTH , 0);
            ctx.lineTo(WIDTH , HEIGHT);
            ctx.lineTo(0 , HEIGHT);
            ctx.lineTo(0 , 0);
            ctx.stroke();

            var fontSize = Math.floor(topMarginOffset / 2.5);
            ctx.font = fontSize + "px Arial";

            ctx.fillText("Option " + (scheduleIndex+1), fontSize*0.5, fontSize*1.2)
            ctx.fillText(schedule.credits + " credits", fontSize*0.5, fontSize*2.2)


            var fontSize = Math.floor(topMarginOffset / 2);
            ctx.font = fontSize + "px Arial";


            //Draw Vertical Lines
            for(let day = 1; day <= DAY_COUNT; day++){
                let xPos = sideMarginOffset + dayWidth*day;

                ctx.fillText(weekDayStrings[day-1], xPos-dayWidth*0.95, topMarginOffset*0.70)

                ctx.beginPath();
                ctx.moveTo(xPos , 0);
                ctx.lineTo(xPos, HEIGHT);
                ctx.stroke();
            }
            var fontSize = Math.floor(dayWidth / 11) + "px";
            ctx.font = fontSize + " Arial";
            ctx.lineWidth = MINOR_LINE_WIDTH;

            var startHourOffset = (Math.ceil(startTime/60)-(startTime/60)) * hourHeight

            console.log( Math.ceil(startTime/60) - (startTime/60));
            console.log(startHourOffset);
            //Draw horizontal lines
            for(let hour = 0; hour < hourCount+1; hour++){
                let yPos = topMarginOffset + hourHeight*hour + startHourOffset;

                if(hour != 0){
                    var time = minsToPrettyTime( (Math.ceil(startTime/60) + (hour-1)) *60);
                    ctx.fillText(time, sideMarginOffset*0.22, yPos-hourHeight+ 2*Math.floor(dayWidth / 11) );
                }

                ctx.beginPath();
                ctx.moveTo(0 , yPos);
                ctx.lineTo(WIDTH, yPos);
                ctx.stroke();
            }

            for(let i = 0; i < schedule.count; i++){
                drawCourse( schedule.courses[i], ctx, dayWidth, hourHeight, sideMarginOffset, topMarginOffset)
            }
        }

    }

    function drawCourse(course, ctx, dayWidth, hourHeight, sideMarginOffset, topMarginOffset){
        
        console.log(course);

        for(let i = 0; i < course.days.length; i++){
            for(let day = 0; day < DAY_COUNT; day++){
                var onDay = daysOverlap(course.days[i], weekDayChars[day])

                if(onDay){
                    var xPos = sideMarginOffset + dayWidth*day;
                    var yPos = topMarginOffset + (hourHeight*(course.starts[i]-startTime)/60);

                    //console.log("Shade: " + course.shade + ", Color: " + course.color);
                    var color = colStrToRGB(course.shade, course.color);

                    ctx.fillStyle = `rgb(
                        ${color[0]},
                        ${color[1]},
                        ${color[2]})`


                    var duration = (course.ends[i] - course.starts[i])/60
                    ctx.fillRect(xPos + MAJOR_LINE_WIDTH/2, yPos, dayWidth - MAJOR_LINE_WIDTH, duration*hourHeight);

                    ctx.fillStyle = findTextColor(color);
                    

                    
                    //var fontSize = Math.floor(hourHeight /4);
                    var fontSize = Math.floor(dayWidth / 11);
                    ctx.font = fontSize + "px Arial";
                
                    
                    var notes = course.notes;
                    var notes2 = ""

                    console.log(ctx.measureText(notes));

                    while(ctx.measureText(notes).width >= dayWidth*.9){
                        notes2 = notes[notes.length-1] + notes2;
                        notes = notes.substr(0, notes.length-1);
                    }

                    ctx.fillText(course.name, xPos + fontSize/2, yPos + fontSize*1.1);
                    ctx.fillText(course.crn + " - " + course.credits, xPos + fontSize/2, yPos + fontSize*2.1);
                    ctx.fillText(notes, xPos + fontSize/2, yPos + fontSize*3.1);
                    ctx.fillText(notes2, xPos + fontSize/2, yPos + fontSize*4.1);
                }
            }
        }
    }

    function daysOverlap(days1, days2){
        for(let i = 0; i < days1.length; i++){
            if (days2.toLowerCase().includes( days1.toLowerCase().charAt(i) ) ){
                return true;
            }
        }
        return false;
    }
    function colStrToRGB(shadeStr, colorIndStr){
        //Select the correct color from this list


        var colors = [ [[212, 21, 21], [150,0,0], [255,128,128]], //Red
                       [[255,128,0], [148,74,0], [255, 175, 94]], //Orange
                       [[255, 230, 0], [224, 183, 0], [255, 255, 92]], //Yellow
                       [ [43,179,43], [83,219,83], [0,130,0]], //Green
                       [[11,230,226], [23,156,153], [161, 255, 252]], //Teal
                       [[0,115,255], [0,78,194], [74,167,255]], //Blue
                       [[162,23,255], [83,0,138], [223,64,255]], //Purple
                       [[128,128,128], [64,64,64], [191,191,191]]] //Grey


        return colors[parseInt(colorIndStr)][parseInt(shadeStr)]
    }
    function findTextColor(color){
        if( (0.299*color[0] + 0.587*color[1] + 0.114*color[2]) > 128 ){
            return "#000000";
        }
        else{
            return "#E0E0E0";
        }
    }  


    var allSchedules = [
        {
            credits: 12,
            count: 2,
            courses: [
                {
                    name: "Math",
                    crn: "12345",
                    credits: 3,
                    days: ["m", "w", "f"],
                    starts: [540, 600, 660],
                    ends: [600, 660, 720],
                    shade: "2",
                    color: [[212, 21, 21], [150,0,0], [255,128,128]],
                    notes: "Room 101"
                },
                {
                    name: "Physics",
                    crn: "54321",
                    credits: 4,
                    days: ["t", "r"],
                    starts: [540, 600],
                    ends: [600, 660],
                    shade: "1",
                    color: "1",
                    notes: "Lab 301"
                },
                // Add more courses as needed
            ]
        },
        {
            credits: 15,
            count: 1,
            courses: [
                {
                    name: "History",
                    crn: "98765",
                    credits: 3,
                    days: ["m", "w", "f"],
                    starts: [720, 780, 840],
                    ends: [780, 840, 900],
                    shade: "1",
                    color: "1",
                    notes: "Room 201"
                },
                // Add more courses as needed
            ]
        },
        // Add more schedules as needed
    ];
    

// Set other necessary variables like startTime, endTime, etc.

// Call drawSchedule function with scheduleIndex
drawSchedule(0); // Draw the first schedule in the array
