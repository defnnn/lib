package main

import (
	_ "embed"
	"encoding/json"

	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/cdktf/cdktf-provider-aws-go/aws/v10/dataawsssoadmininstances"
	"github.com/cdktf/cdktf-provider-aws-go/aws/v10/identitystoregroup"
	"github.com/cdktf/cdktf-provider-aws-go/aws/v10/identitystoregroupmembership"
	"github.com/cdktf/cdktf-provider-aws-go/aws/v10/identitystoreuser"
	"github.com/cdktf/cdktf-provider-aws-go/aws/v10/organizationsaccount"
	"github.com/cdktf/cdktf-provider-aws-go/aws/v10/organizationsorganization"
	aws "github.com/cdktf/cdktf-provider-aws-go/aws/v10/provider"
	"github.com/cdktf/cdktf-provider-aws-go/aws/v10/ssoadminaccountassignment"
	"github.com/cdktf/cdktf-provider-aws-go/aws/v10/ssoadminmanagedpolicyattachment"
	"github.com/cdktf/cdktf-provider-aws-go/aws/v10/ssoadminpermissionset"

	tfe "github.com/cdktf/cdktf-provider-tfe-go/tfe/v3/provider"
	"github.com/cdktf/cdktf-provider-tfe-go/tfe/v3/workspace"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

//go:embed schema/aws.cue
var aws_schema_cue string

type TerraformCloud struct {
	Organization string `json:"organization"`
	Workspace    string `json:"workspace"`
}

type AwsAdmin struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type AwsOrganization struct {
	Name     string     `json:"name"`
	Region   string     `json:"region"`
	Prefix   string     `json:"prefix"`
	Domain   string     `json:"domain"`
	Accounts []string   `json:"accounts"`
	Admins   []AwsAdmin `json:"admins"`
}

type AwsProps struct {
	Terraform     TerraformCloud             `json:"terraform"`
	Organizations map[string]AwsOrganization `json:"organizations"`
}

// alias
func js(s string) *string {
	return jsii.String(s)
}

func TfcOrganizationWorkspacesStack(scope constructs.Construct, id string) cdktf.TerraformStack {
	stack := cdktf.NewTerraformStack(scope, js(id))

	tfe.NewTfeProvider(stack, js("tfe"), &tfe.TfeProviderConfig{
		Hostname: js("app.terraform.io"),
	})

	return stack
}

func AwsOrganizationStack(scope constructs.Construct, org *AwsOrganization) cdktf.TerraformStack {
	stack := cdktf.NewTerraformStack(scope, js(org.Name))

	aws.NewAwsProvider(stack,
		js("aws"), &aws.AwsProviderConfig{
			Region: js(org.Region),
		})

	organizationsorganization.NewOrganizationsOrganization(stack,
		js("organization"),
		&organizationsorganization.OrganizationsOrganizationConfig{
			FeatureSet: js("ALL"),
			EnabledPolicyTypes: &[]*string{
				js("SERVICE_CONTROL_POLICY"),
				js("TAG_POLICY")},
			AwsServiceAccessPrincipals: &[]*string{
				js("cloudtrail.amazonaws.com"),
				js("config.amazonaws.com"),
				js("ram.amazonaws.com"),
				js("ssm.amazonaws.com"),
				js("sso.amazonaws.com"),
				js("tagpolicies.tag.amazonaws.com")},
		})
	// Lookup pre-enabled AWS SSO instance
	ssoadmin_instance := dataawsssoadmininstances.NewDataAwsSsoadminInstances(stack,
		js("sso_instance"),
		&dataawsssoadmininstances.DataAwsSsoadminInstancesConfig{})

	ssoadmin_instance_arn := cdktf.NewTerraformLocal(stack,
		js("sso_instance_arn"),
		ssoadmin_instance.Arns())

	ssoadmin_permission_set := ssoadminpermissionset.NewSsoadminPermissionSet(stack,
		js("admin_sso_permission_set"),
		&ssoadminpermissionset.SsoadminPermissionSetConfig{
			Name:            js("Administrator"),
			InstanceArn:     js(cdktf.Fn_Element(ssoadmin_instance_arn.Expression(), jsii.Number(0)).(string)),
			SessionDuration: js("PT2H"),
			Tags:            &map[string]*string{"ManagedBy": js("Terraform")},
		})

	sso_permission_set_admin := ssoadminmanagedpolicyattachment.NewSsoadminManagedPolicyAttachment(stack,
		js("admin_sso_managed_policy_attachment"),
		&ssoadminmanagedpolicyattachment.SsoadminManagedPolicyAttachmentConfig{
			InstanceArn:      ssoadmin_permission_set.InstanceArn(),
			PermissionSetArn: ssoadmin_permission_set.Arn(),
			ManagedPolicyArn: js("arn:aws:iam::aws:policy/AdministratorAccess"),
		})

	ssoadmin_instance_isid := cdktf.NewTerraformLocal(stack,
		js("sso_instance_isid"),
		ssoadmin_instance.IdentityStoreIds())

	// Create Administrators group
	identitystore_group := identitystoregroup.NewIdentitystoreGroup(stack,
		js("administrators_sso_group"),
		&identitystoregroup.IdentitystoreGroupConfig{
			DisplayName:     js("Administrators"),
			IdentityStoreId: js(cdktf.Fn_Element(ssoadmin_instance_isid.Expression(), jsii.Number(0)).(string)),
		})

	// Create initial users in the Administrators group
	for _, adm := range org.Admins {
		identitystore_user := identitystoreuser.NewIdentitystoreUser(stack,
			js(fmt.Sprintf("admin_sso_user_%s", adm.Name)),
			&identitystoreuser.IdentitystoreUserConfig{
				DisplayName: js(adm.Name),
				UserName:    js(adm.Name),
				Name: &identitystoreuser.IdentitystoreUserName{
					GivenName:  js(adm.Name),
					FamilyName: js(adm.Name),
				},
				Emails: &identitystoreuser.IdentitystoreUserEmails{
					Primary: jsii.Bool(true),
					Type:    js("work"),
					Value:   js(adm.Email),
				},
				IdentityStoreId: js(cdktf.Fn_Element(ssoadmin_instance_isid.Expression(), jsii.Number(0)).(string)),
			})

		identitystoregroupmembership.NewIdentitystoreGroupMembership(stack,
			js(fmt.Sprintf("admin_sso_user_%s_membership", adm.Name)),
			&identitystoregroupmembership.IdentitystoreGroupMembershipConfig{
				MemberId:        identitystore_user.UserId(),
				GroupId:         identitystore_group.GroupId(),
				IdentityStoreId: js(cdktf.Fn_Element(ssoadmin_instance_isid.Expression(), jsii.Number(0)).(string)),
			})
	}

	// The master account (named "org") must be imported.
	for _, acct := range append(org.Accounts, []string{org.Name}...) {
		// Create the organization account
		var organizations_account_config organizationsaccount.OrganizationsAccountConfig

		if acct == org.Name {
			// The master organization account can't set
			// iam_user_access_to_billing, role_name
			organizations_account_config = organizationsaccount.OrganizationsAccountConfig{
				Name:  js(acct),
				Email: js(fmt.Sprintf("%s%s@%s", org.Prefix, org.Name, org.Domain)),
				Tags:  &map[string]*string{"ManagedBy": js("Terraform")},
			}
		} else {
			// Organization account
			organizations_account_config = organizationsaccount.OrganizationsAccountConfig{
				Name:                   js(acct),
				Email:                  js(fmt.Sprintf("%s%s+%s@%s", org.Prefix, org.Name, acct, org.Domain)),
				Tags:                   &map[string]*string{"ManagedBy": js("Terraform")},
				IamUserAccessToBilling: js("ALLOW"),
				RoleName:               js("OrganizationAccountAccessRole"),
			}
		}

		organizations_account := organizationsaccount.NewOrganizationsAccount(stack,
			js(acct),
			&organizations_account_config)

		// Organization accounts grant Administrator permission set to the Administrators group
		ssoadminaccountassignment.NewSsoadminAccountAssignment(stack,
			js(fmt.Sprintf("%s_admin_sso_account_assignment", acct)),
			&ssoadminaccountassignment.SsoadminAccountAssignmentConfig{
				InstanceArn:      sso_permission_set_admin.InstanceArn(),
				PermissionSetArn: sso_permission_set_admin.PermissionSetArn(),
				PrincipalId:      identitystore_group.GroupId(),
				PrincipalType:    js("GROUP"),
				TargetId:         organizations_account.Id(),
				TargetType:       js("AWS_ACCOUNT"),
			})
	}

	return stack
}

func LoadUserAwsProps() AwsProps {
	ctx := cuecontext.New()

	user_schema := ctx.CompileString(aws_schema_cue)

	user_input_instance := load.Instances([]string{"."}, nil)[0]
	user_input := ctx.BuildInstance(user_input_instance)

	user_schema.Unify(user_input)

	var aws_props AwsProps
	user_input.LookupPath(cue.ParsePath("input")).Decode(&aws_props)

	return aws_props
}

func QueueAwsProps(hostport string) {
	c, err := client.Dial(client.Options{HostPort: hostport})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("aws_organizations_%d", time.Now().UnixMilli()),
		TaskQueue: "aws-organizations",
	}

	aws_props := LoadUserAwsProps()

	fmt.Printf("%v\n", aws_props)

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, AwsOrganizationsWorkflow, aws_props)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Synchronously wait for the workflow completion.
	var result = make(map[string]any)
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}
	log.Printf("Workflow result:\n%v\n", result)
}

func AwsOrganizationsWorker(hostport string) {
	c, err := client.Dial(client.Options{HostPort: hostport})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "aws-organizations", worker.Options{})

	w.RegisterWorkflow(AwsOrganizationsWorkflow)
	w.RegisterActivity(AwsOrganizationsActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

func AwsOrganizationsWorkflow(ctx workflow.Context, aws_props AwsProps) (map[string]any, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)

	var result = make(map[string]any)
	err := workflow.ExecuteActivity(ctx, AwsOrganizationsActivity, &aws_props).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return nil, err
	}

	return result, nil
}

func AwsOrganizationsActivity(ctx context.Context, aws_props AwsProps) (map[string]any, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity")

	// Our app manages the tfc workspaces, aws organizations plus their accounts
	app := cdktf.NewApp(nil)

	workspaces := TfcOrganizationWorkspacesStack(app, aws_props.Terraform.Workspace)

	cdktf.NewCloudBackend(workspaces, &cdktf.CloudBackendProps{
		Hostname:     js("app.terraform.io"),
		Organization: js(aws_props.Terraform.Organization),
		Workspaces:   cdktf.NewNamedCloudWorkspace(js("workspaces")),
	})

	for _, org := range aws_props.Organizations {
		// Create a tfc workspace for each stack
		workspace.NewWorkspace(workspaces, js(org.Name), &workspace.WorkspaceConfig{
			Name:                js(org.Name),
			Organization:        js(aws_props.Terraform.Organization),
			ExecutionMode:       js("local"),
			FileTriggersEnabled: false,
			QueueAllRuns:        false,
			SpeculativeEnabled:  false,
		})

		// Create the aws organization + accounts stack
		aws_org_stack := AwsOrganizationStack(app, &org)
		cdktf.NewCloudBackend(aws_org_stack, &cdktf.CloudBackendProps{
			Hostname:     js("app.terraform.io"),
			Organization: js(aws_props.Terraform.Organization),
			Workspaces:   cdktf.NewNamedCloudWorkspace(js(org.Name)),
		})
	}

	// Emit cdk.tf.json
	app.Synth()

	// Build map of stack and synthesized tf config
	var synth = make(map[string]any)

	f, _ := os.Open("cdktf.out/stacks/")
	files, _ := f.Readdir(0)
	for _, file := range files {
		dat, _ := os.ReadFile(fmt.Sprintf("cdktf.out/stacks/%s/cdk.tf.json", file.Name()))
		stack := make(map[string]any)
		json.Unmarshal([]byte(dat), &stack)
		synth[file.Name()] = stack
	}

	return synth, nil
}

func main() {
	hostport := os.Args[1]

	if len(os.Args) > 2 && os.Args[2] == "queue" {
		QueueAwsProps(hostport)
	} else {
		AwsOrganizationsWorker(hostport)
	}
}
