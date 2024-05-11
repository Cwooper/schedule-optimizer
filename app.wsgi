# app.wsgi
import sys
sys.path.insert(0, '/var/www/schedule-optimizer')

from app import app as application
