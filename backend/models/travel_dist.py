# travel_dist.py
# Built-in objects
# TODO remove this as an object and move it into Schedule
from backend.models.Schedule import Schedule
from backend.models.Course import Course

import re
import dijkstra # pip3 install dijkstra | 

building_pattern = r'([A-Z]{2})\w'

class travel_dist:
    def __init__(self, schedule: Schedule):
        self.schedule = schedule
        # filter out only the days in schedule.courses

    def get_dist_weight(self):
        days_dict = {
            "M": [],
            "T": [],
            "W": [],
            "R": [],
            "F": []
        }
        for day in "MTWRF":
            for course in self.schedule.courses:
                if course.days.contains(day):
                    days_dict[day].append(course)
        
        # An ordered dictionary of buildings to travel to for each day
        building_dict = {
            "M": [],
            "T": [],
            "W": [],
            "R": [],
            "F": []
        }
        
        for day, _ in days_dict.items():
            # Sort the classes based on value.course.start_time
            days_dict[day].sort(key=lambda x: x.start_time)
            for course in days_dict[day]:
                # Pull out the buildings if they were found
                building_match = re.match(building_pattern, course.room)

                if building_match:
                    building = building_match.group(1)
                else:
                    building = None
                building_dict[day].append(building)

        walk_time = 0
        populated_lists_count = sum(1 for lst in building_dict.values() if lst)
        for day, buildings in building_dict.items():
            for i in range(len(buildings)-1):
                walk_time += get_distance(buildings[i], buildings[i + 1])

        return walk_time / populated_lists_count
        
        



def get_distance(start, end):
    nodes = ["SV", "AW", "AI", "CF", "ES", "PH", "IS", "AH", "BI", "CB", "ET",
             "CB", "SP", "AA", "FI", "AA", "CV", "MH", "BH", "FR", "HU", "HH",
             "WL", "OM", "PA", "VU", "SL"]

    graph = dijkstra.Graph()

    # graph edge, start bldg, end bldg, feet between
    graph.add_edge("SV", "AW", 425)
    graph.add_edge("SV", "PH", 780)
    graph.add_edge("AW", "PH", 510)
    graph.add_edge("AW", "CF", 320)
    graph.add_edge("AW", "ES", 550)
    graph.add_edge("CF", "ES", 330)
    graph.add_edge("PH", "ES", 220)
    graph.add_edge("PH", "IS", 200)
    graph.add_edge("PH", "BI", 225)
    graph.add_edge("PH", "AH", 230)
    graph.add_edge("ES", "AH", 139)
    graph.add_edge("BI", "CB", 215)
    graph.add_edge("BI", "ET", 300)
    graph.add_edge("BI", "AH", 200)
    graph.add_edge("AH", "ET", 230)
    graph.add_edge("ET", "CB", 130)
    graph.add_edge("ET", "FI", 290)
    graph.add_edge("FI", "CV", 400)
    graph.add_edge("CV", "AA", 180)
    graph.add_edge("CV", "BH", 150)
    graph.add_edge("CV", "MH", 200)
    graph.add_edge("CV", "SL", 241)
    graph.add_edge("MH", "FR", 280)
    graph.add_edge("MH", "HH", 360)
    graph.add_edge("MH", "HU", 310)
    graph.add_edge("HU", "HH", 160)
    graph.add_edge("HU", "BH", 260)
    graph.add_edge("HU", "OM", 180)
    graph.add_edge("HU", "WL", 150)
    graph.add_edge("OM", "VU", 453)
    graph.add_edge("WL", "VU", 280)
    graph.add_edge("VU", "PA", 330)
    graph.add_edge("HH", "PA", 220)
    graph.add_edge("PA", "WL", 200)
    
    all_dist = dijkstra.DijkstraSPF(graph, start)

    dist = all_dist.get_path(end)

    minutes = dist / 285    # Average walk speed is 285 feet per minute

    return minutes

