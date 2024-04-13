
class Course:
    def __init__(self, name: str, days: str, start: int, end: int):
        self.name = name
        self.days = days
        self.start = start
        self.end = end
    
    def __repr__(self):
        return f'{self.name} {self.days} {self.start} - {self.end}'
    
    def conflicts(self, other) -> bool:
        if self.name == other.name:
            return True
        for day in self.days:
            if day in other.days:
                if (((self.start >= other.start) and (self.start <= other.end)) or
                     ((self.end >= other.start) and (self.end <= other.end))):
                    return True
        return False
    