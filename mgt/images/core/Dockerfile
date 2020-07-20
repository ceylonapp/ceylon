FROM python:3

# set a directory for the app
WORKDIR /usr/src/app

# copy all the files to the container
COPY mgt/bases/core .
COPY mgt/images/core .

# install dependencies
RUN pip install --no-cache-dir -r requirements.txt

# define the port number the container should expose
EXPOSE 5000

RUN ls /usr/src/app

# run the command
CMD ["python", "./app.py"]