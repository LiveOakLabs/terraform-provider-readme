# API specifications can be imported using their ID.
# Importing API specifications is limited due to the behavior of the API
# registry and associating a specification with its definition. When importing,
# Terraform will replace the remote definition on its next run, regardless if it
# differs from the local definition.
terraform import readme_api_specification.example 639fd743a9690100813a13fd