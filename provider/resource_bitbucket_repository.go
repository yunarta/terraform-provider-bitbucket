package provider

import (
	"context"
	"fmt"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bitbucket"
	"github.com/yunarta/terraform-provider-commons/util"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	_ resource.Resource                = &RepositoryResource{}
	_ resource.ResourceWithConfigure   = &RepositoryResource{}
	_ resource.ResourceWithImportState = &RepositoryResource{}
	_ RepositoryPermissionReceiver     = &RepositoryResource{}
	_ ConfigurableReceiver             = &RepositoryResource{}
)

func NewRepositoryResource() resource.Resource {
	return &RepositoryResource{}
}

type RepositoryResource struct {
	config BitbucketProviderConfig
	client *bitbucket.Client
}

func (receiver *RepositoryResource) getClient() *bitbucket.Client {
	return receiver.client
}

func (receiver *RepositoryResource) setConfig(config BitbucketProviderConfig, client *bitbucket.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *RepositoryResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_repository"
}

type createSlug struct {
}

func (r createSlug) Description(ctx context.Context) string {
	return "If the value of this attribute changes, Terraform will destroy and recreate the resource."
}

func (r createSlug) MarkdownDescription(ctx context.Context) string {
	return "If the value of this attribute changes, Terraform will destroy and recreate the resource."
}

func (r createSlug) PlanModifyString(ctx context.Context, request planmodifier.StringRequest, response *planmodifier.StringResponse) {
	var repoName types.String
	request.Plan.GetAttribute(ctx, path.Root("name"), &repoName)

	// Convert the string to lowercase
	slug := strings.ToLower(repoName.ValueString())

	// Reduce consecutive spaces to a single space
	slug = regexp.MustCompile("\\s+").ReplaceAllString(slug, " ")

	// Replace spaces with a single hyphen
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove any characters that are not alphanumeric, hyphens, underscores, or dots
	reg := regexp.MustCompile("[^a-z0-9-_.]")
	slug = reg.ReplaceAllString(slug, "")

	response.PlanValue = types.StringValue(slug)
}

func (receiver *RepositoryResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"retain_on_delete": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"archive_on_delete": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"slug": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					&createSlug{},
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					util.ReplaceIfStringDiff(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9\-\s_.]*$`),
						"must start with a letter or number and may contain spaces, hyphens, underscores, and periods",
					),
				},
				Description: "Repository name",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Repository description",
			},
			"project": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					util.ReplaceIfStringDiff(),
				},
			},
			"readme": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"path": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"assignment_version": schema.StringAttribute{
				Optional: true,
			},
			"computed_users":  ComputedAssignmentSchema,
			"computed_groups": ComputedAssignmentSchema,
		},
		Blocks: map[string]schema.Block{
			"assignments": AssignmentSchema("REPO_ADMIN", "REPO_READ", "REPO_WRITE"),
		},
	}
}

func (receiver *RepositoryResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	ConfigureResource(receiver, ctx, request, response)
}

func (receiver *RepositoryResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var (
		plan  RepositoryModel
		diags diag.Diagnostics
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repository, err := receiver.client.RepositoryService().Create(plan.Project.ValueString(), bitbucket.CreateRepo{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	})
	if util.TestError(&response.Diagnostics, err, errorFailedToCreateRepository) {
		return
	}

	plan.Slug = types.StringValue(repository.Slug)

	if util.TestDiagnostics(
		&response.Diagnostics,
		response.State.SetAttribute(ctx, path.Root("id"), types.StringValue(strconv.Itoa(repository.ID))),
		response.State.SetAttribute(ctx, path.Root("project"), types.StringValue(repository.Project.Key)),
		response.State.SetAttribute(ctx, path.Root("slug"), types.StringValue(repository.Slug)),
	) {
		return
	}

	computation, diags := CreateRepositoryAssignments(ctx, receiver, plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryModel := NewRepositoryModel(repository, plan, computation)

	diags = response.State.Set(ctx, repositoryModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
	if !plan.Readme.IsNull() {
		_, err = receiver.client.RepositoryService().Initialize(
			plan.Project.ValueString(),
			plan.Name.ValueString(),
			plan.Readme.ValueString(),
		)
		if util.TestError(&response.Diagnostics, err, errorFailedToInitializeRepository) {
			return
		}
	} else if !plan.Path.IsNull() {
		// plan.Project need top be lower case

		repo, err := git.Init(memory.NewStorage(), memfs.New())
		if util.TestError(&response.Diagnostics, err, errorFailedToInitializeRepository) {
			return
		}

		worktree, err := repo.Worktree()

		// copy filesystem to worktree recursively
		err = filepath.Walk(plan.Path.ValueString(), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil // skip directories
			}

			srcFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			relativePath, err := filepath.Rel(plan.Path.ValueString(), path)
			if err != nil {
				return err
			}

			destPath := filepath.Join(worktree.Filesystem.Root(), relativePath)
			destFile, err := worktree.Filesystem.Create(destPath)
			if err != nil {
				return err
			}

			_, err = io.Copy(destFile, srcFile)
			if err != nil {
				return err
			}

			return destFile.Close()
		})

		if util.TestError(&response.Diagnostics, err, errorFailedToInitializeRepository) {
			return
		}

		_, err = worktree.Add(".")
		if util.TestError(&response.Diagnostics, err, errorFailedToInitializeRepository) {
			return
		}

		_, err = worktree.Commit("Initial Commit", &git.CommitOptions{
			Author: &object.Signature{
				Name:  receiver.config.Author.Name.ValueString(),
				Email: receiver.config.Author.Email.ValueString(),
				When:  time.Now(),
			},
		})
		if util.TestError(&response.Diagnostics, err, errorFailedToInitializeRepository) {
			return
		}

		_, err = repo.CreateRemote(&config.RemoteConfig{
			Name: "origin",
			URLs: []string{
				fmt.Sprintf("%s/scm/%s/%s.git",
					receiver.config.Bitbucket.EndPoint.ValueString(),
					strings.ToLower(plan.Project.ValueString()),
					repository.Slug),
			}, // replace with your remote repo URL
		})
		if util.TestError(&response.Diagnostics, err, errorFailedToInitializeRepository) {
			return
		}

		var auth transport.AuthMethod
		if receiver.config.Bitbucket.Token.IsNull() {
			auth = &http.BasicAuth{
				Username: receiver.config.Bitbucket.Username.ValueString(),
				Password: receiver.config.Bitbucket.Password.ValueString(),
			}
		} else {
			auth = &http.BasicAuth{
				Username: receiver.config.Bitbucket.Username.ValueString(),
				Password: receiver.config.Bitbucket.Token.ValueString(),
			}
		}
		err = repo.Push(&git.PushOptions{
			Auth: auth,
			//RefSpecs: []config.RefSpec{
			//	"refs/heads/master:refs/heads/master",
			//},
		})
		if util.TestError(&response.Diagnostics, err, errorFailedToInitializeRepository) {
			return
		}
	}

}

func (receiver *RepositoryResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var (
		diags diag.Diagnostics
		err   error

		state RepositoryModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repository, err := receiver.client.RepositoryService().Read(state.Project.ValueString(), state.Slug.ValueString())
	if util.TestError(&response.Diagnostics, err, errorFailedToCreateRepository) {
		return
	}

	computation, diags := ComputeRepositoryAssignments(ctx, receiver, state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryModel := NewRepositoryModel(repository, state, computation)

	diags = response.State.Set(ctx, repositoryModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *RepositoryResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		diags diag.Diagnostics
		err   error

		plan, state RepositoryModel
	)

	if util.TestDiagnostics(&response.Diagnostics,
		request.Plan.Get(ctx, &plan),
		request.State.Get(ctx, &state)) {
		return
	}

	repository, err := receiver.client.RepositoryService().Update(
		state.Project.ValueString(),
		state.Slug.ValueString(),
		plan.Description.ValueString(),
	)
	if util.TestError(&response.Diagnostics, err, errorFailedToUpdateRepository) {
		return
	}

	forceUpdate := !plan.AssignmentVersion.Equal(state.AssignmentVersion)
	computation, diags := UpdateRepositoryAssignments(ctx, receiver, plan, state, forceUpdate)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryModel := NewRepositoryModel(repository, plan, computation)

	diags = response.State.Set(ctx, repositoryModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *RepositoryResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var (
		diags diag.Diagnostics
		err   error

		state RepositoryModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	if !state.RetainOnDelete.ValueBool() {
		diags = DeleteRepositoryAssignments(ctx, receiver, state)
		if util.TestDiagnostic(&response.Diagnostics, diags) {
			return
		}

		if state.ArchiveOnDelete.ValueBool() {

			currentTime := time.Now().Format("2006-01-02-15-04-05")
			err = receiver.client.RepositoryService().Rename(
				state.Project.ValueString(),
				state.Slug.ValueString(),
				fmt.Sprintf("%s-archived-at-%s", state.Slug.ValueString(), currentTime),
			)
		} else {
			err = receiver.client.RepositoryService().Delete(
				state.Project.ValueString(),
				state.Slug.ValueString(),
			)
		}

		if util.TestError(&response.Diagnostics, err, errorFailedToDeleteRepository) {
			return
		}
	}

	response.State.RemoveResource(ctx)
}

func (receiver *RepositoryResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	slug := strings.Split(request.ID, "/")
	diags := response.State.Set(ctx, &RepositoryModel{
		Project:        types.StringValue(slug[0]),
		Slug:           types.StringValue(slug[1]),
		Assignments:    types.ListNull(assignmentType),
		ComputedUsers:  types.ListNull(computedAssignmentType),
		ComputedGroups: types.ListNull(computedAssignmentType),
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}
