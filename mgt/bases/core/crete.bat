@ECHO OFF
ECHO Welcome to ceylon console
python -m venv venv
.\venv\Scripts\activate
python --version
pip install -r .\requirements.txt
