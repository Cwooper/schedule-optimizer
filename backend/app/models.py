from django.db import models

class Class(models.Model):
    name = models.CharField(max_length=100)     # Name of the class
    days = models.CharField(max_length=7)       # Ex: "TR" for Tuesday, Thursday
    start_time = models.TimeField()             # Class start time
    end_time = models.TimeField()               # Class end time

    # Possible Future Additions:

    # Class CRN
    # Class Term
    # Class Credits
    # Instructor first name
    # Instructor Last Name
    # Meeting Place

    # Lab Room
    # Lab start time
    # Lab end time