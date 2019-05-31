# Upgrading to Terraform 0.12

The best way to proceed here is to download the latest version of the OpenAPI Terraform provider plugin that has support for
the Terraform 0.12 SDK, at the time of writing it's v0.14.0.

## What if I already have infrastructure managed via Terraform using OpenAPI plugin version < v0.14.0

In that case, you should be able to upgrade the terraform binary to 0.12 and also the OpenAPI Terraform provider plugin (>=v0.14.0)

### What should I expect?

Once provisioned with the latest binary versions, run ```terraform plan``` on your configuration. One of the following
use cases will apply:

#### No brainier, your config is already compatible

The plan will show:
 
````
No changes. Infrastructure is up-to-date.
````

In this case you don't have to do anything else.

#### Your configuration is not compatible with Terraform 0.12.

Terraform will present you the different config that is not compatible. For instance:

````
$ terraform plan

Error: Unsupported argument

  on main.tf line 60, in resource "swaggercodegen_cdn_v1" "my_cdn2":
  60:   array_of_objects_example = [

An argument named "array_of_objects_example" is not expected here. Did you
mean to define a block of type "array_of_objects_example"?
````

In the case above, Terraform is notifying us that there is a property that is not compatible with Terraform 0.12.

Luckily, Terraform 0.12 supports a new command that updates your current non compatible configuration into a compatible 
Terraform 0.12 configuration. The command is called [0.12upgrade](https://www.terraform.io/docs/commands/0.12upgrade.html). Below
is an example of an output:

````
$ terraform 0.12upgrade

This command will rewrite the configuration files in the given directory so
that they use the new syntax features from Terraform v0.12, and will identify
any constructs that may need to be adjusted for correct operation with
Terraform v0.12.

We recommend using this command in a clean version control work tree, so that
you can easily see the proposed changes as a diff against the latest commit.
If you have uncommited changes already present, we recommend aborting this
command and dealing with them before running this command again.

Would you like to upgrade the module in the current directory?
  Only 'yes' will be accepted to confirm.

  Enter a value: yes

-----------------------------------------------------------------------------

Warning: Approximate migration of invalid block type assignment

  on main.tf line 61, in resource "swaggercodegen_cdn_v1" "my_cdn2":
  61:     "${swaggercodegen_cdn_v1.my_cdn.array_of_objects_example[0]}",

In swaggercodegen_cdn_v1.my_cdn2 the name "array_of_objects_example" is a
nested block type, but this configuration is exploiting some missing
validation rules from Terraform v0.11 and prior to trick Terraform into
creating blocks dynamically.

This has been upgraded to use the new Terraform v0.12 dynamic blocks feature,
but since the upgrade tool cannot predict which map keys will be present a
fully-comprehensive set has been generated.


Warning: Approximate migration of invalid block type assignment

  on main.tf line 62, in resource "swaggercodegen_cdn_v1" "my_cdn2":
  62:     "${swaggercodegen_cdn_v1.my_cdn.array_of_objects_example[1]}",

In swaggercodegen_cdn_v1.my_cdn2 the name "array_of_objects_example" is a
nested block type, but this configuration is exploiting some missing
validation rules from Terraform v0.11 and prior to trick Terraform into
creating blocks dynamically.

This has been upgraded to use the new Terraform v0.12 dynamic blocks feature,
but since the upgrade tool cannot predict which map keys will be present a
fully-comprehensive set has been generated.

-----------------------------------------------------------------------------

Upgrade complete!

The configuration files were upgraded successfully. Use your version control
system to review the proposed changes, make any necessary adjustments, and
then commit.

Some warnings were generated during the upgrade, as shown above. These
indicate situations where Terraform could not decide on an appropriate course
of action without further human input.

Where possible, these have also been marked with TF-UPGRADE-TODO comments to
mark the locations where a decision must be made. After reviewing and adjusting
these, manually remove the TF-UPGRADE-TODO comment before continuing.
````

After executing the above command, you should expect some changes in your terraform configuration file. In the case above,
these changes where related to how arrays ob objects are represented in Terraform 0.12. The snipped of code
updated can be seen below:

````
  dynamic "array_of_objects_example" {
    for_each = [swaggercodegen_cdn_v1.my_cdn.array_of_objects_example[0]]
    content {
      # TF-UPGRADE-TODO: The automatic upgrade tool can't predict
      # which keys might be set in maps assigned here, so it has
      # produced a comprehensive set here. Consider simplifying
      # this after confirming which keys can be set in practice.

      origin_port = lookup(array_of_objects_example.value, "origin_port", null)
      protocol    = lookup(array_of_objects_example.value, "protocol", null)
    }
  }
  dynamic "array_of_objects_example" {
    for_each = [swaggercodegen_cdn_v1.my_cdn.array_of_objects_example[1]]
    content {
      # TF-UPGRADE-TODO: The automatic upgrade tool can't predict
      # which keys might be set in maps assigned here, so it has
      # produced a comprehensive set here. Consider simplifying
      # this after confirming which keys can be set in practice.

      origin_port = lookup(array_of_objects_example.value, "origin_port", null)
      protocol    = lookup(array_of_objects_example.value, "protocol", null)
    }
  }
````

And Voilà! We now have a fully Terraform 0.12 configuration file. The expectation at this point, is that only the configuration
should have changed. Executing terraform plan should not see any diffs at all.

````
$ terraform plan
Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.

swaggercodegen_lbs_v1.my_lb: Refreshing state... [id=c438fc1b-a85a-48eb-afc6-2e6d719d3c82]
swaggercodegen_cdn_v1.my_cdn: Refreshing state... [id=403cd232-89cd-4586-9f22-1b37786e698f]
swaggercodegen_cdn_v1.my_cdn2: Refreshing state... [id=846c997b-ba02-4e1c-984c-e8ff465fa76b]

------------------------------------------------------------------------

No changes. Infrastructure is up-to-date.

This means that Terraform did not detect any differences between your
configuration and real physical resources that exist. As a result, no
actions need to be performed.
````