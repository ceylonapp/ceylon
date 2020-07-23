FROM python:3.7.8-buster

# set a directory for the app
WORKDIR /usr/src/app

# copy all the files to the container
COPY . .

RUN ls .
# install dependencies
RUN pip install --no-cache-dir -r config/requirements.txt

RUN ls .
#RUN chmod +x ./run.sh
# run the command
CMD ["/bin/sh","run.sh"]
{{- range $val := .Expose }}
EXPOSE {{$val}}
{{end}}