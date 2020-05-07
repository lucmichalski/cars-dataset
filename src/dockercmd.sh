#/bin/bash
# executed on docker-image (aws task) startup

#AWS_REGION=eu-central-1

######

# read sensitive information from AWS EC2 parameter store
# this uses the 'ecsInstance' role which needs access rights on SSM::Get-Parameters
# (1) retrieve data from aws ssm store
# (2) extract value from json using jquery (-r is 'raw', no parentheses and proper new lines)

#mkdir pem

# NGINX
#aws ssm get-parameters --names $SSM_ID_NGINX_CACERT --no-with-decryption --region $AWS_REGION --output json | jq -r '.Parameters[0] | .Value' > ./pem/nginx.cacert.pem
#aws ssm get-parameters --names $SSM_ID_NGINX_CERT --no-with-decryption --region $AWS_REGION --output json | jq -r '.Parameters[0] | .Value' > ./pem/nginx.cert.pem
#aws ssm get-parameters --names $SSM_ID_NGINX_KEY --with-decryption --region $AWS_REGION --output json | jq -r '.Parameters[0] | .Value' > ./pem/nginx.key.pem

# CASSANDRA
#aws ssm get-parameters --names $SSM_ID_CASSANDRA_CACERT --no-with-decryption --region $AWS_REGION --output json | jq -r '.Parameters[0] | .Value' > ./pem/cassandra.cacert.pem
#aws ssm get-parameters --names $SSM_ID_CASSANDRA_CERT --no-with-decryption --region $AWS_REGION --output json | jq -r '.Parameters[0] | .Value' > ./pem/cassandra.cert.pem
#aws ssm get-parameters --names $SSM_ID_CASSANDRA_KEY --with-decryption --region $AWS_REGION --output json | jq -r '.Parameters[0] | .Value' > ./pem/cassandra.key.pem

# NATS
#aws ssm get-parameters --names $SSM_ID_NATS_CACERT --no-with-decryption --region $AWS_REGION --output json | jq -r '.Parameters[0] | .Value' > ./pem/nats.cacert.pem
#aws ssm get-parameters --names $SSM_ID_NATS_CERT --no-with-decryption --region $AWS_REGION --output json | jq -r '.Parameters[0] | .Value' > ./pem/nats.cert.pem
#aws ssm get-parameters --names $SSM_ID_NATS_KEY --with-decryption --region $AWS_REGION --output json | jq -r '.Parameters[0] | .Value' > ./pem/nats.key.pem

######

# start supervisor
supervisord -c /etc/supervisor.conf
