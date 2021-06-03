// *** WARNING: this file was generated by the Pulumi SDK Generator. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

package xyz

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Todo struct {
	pulumi.CustomResourceState

	Completed pulumi.BoolOutput   `pulumi:"completed"`
	Order     pulumi.IntOutput    `pulumi:"order"`
	Title     pulumi.StringOutput `pulumi:"title"`
	Url       pulumi.StringOutput `pulumi:"url"`
}

// NewTodo registers a new resource with the given unique name, arguments, and options.
func NewTodo(ctx *pulumi.Context,
	name string, args *TodoArgs, opts ...pulumi.ResourceOption) (*Todo, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.Title == nil {
		return nil, errors.New("invalid value for required argument 'Title'")
	}
	var resource Todo
	err := ctx.RegisterResource("xyz:index:Todo", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// GetTodo gets an existing Todo resource's state with the given name, ID, and optional
// state properties that are used to uniquely qualify the lookup (nil if not required).
func GetTodo(ctx *pulumi.Context,
	name string, id pulumi.IDInput, state *TodoState, opts ...pulumi.ResourceOption) (*Todo, error) {
	var resource Todo
	err := ctx.ReadResource("xyz:index:Todo", name, id, state, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

// Input properties used for looking up and filtering Todo resources.
type todoState struct {
	Completed *bool   `pulumi:"completed"`
	Order     *int    `pulumi:"order"`
	Title     *string `pulumi:"title"`
	Url       *string `pulumi:"url"`
}

type TodoState struct {
	Completed pulumi.BoolPtrInput
	Order     pulumi.IntPtrInput
	Title     pulumi.StringPtrInput
	Url       pulumi.StringPtrInput
}

func (TodoState) ElementType() reflect.Type {
	return reflect.TypeOf((*todoState)(nil)).Elem()
}

type todoArgs struct {
	Completed *bool   `pulumi:"completed"`
	Order     *int    `pulumi:"order"`
	Title     string  `pulumi:"title"`
	Url       *string `pulumi:"url"`
}

// The set of arguments for constructing a Todo resource.
type TodoArgs struct {
	Completed pulumi.BoolPtrInput
	Order     pulumi.IntPtrInput
	Title     pulumi.StringInput
	Url       pulumi.StringPtrInput
}

func (TodoArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*todoArgs)(nil)).Elem()
}

type TodoInput interface {
	pulumi.Input

	ToTodoOutput() TodoOutput
	ToTodoOutputWithContext(ctx context.Context) TodoOutput
}

func (*Todo) ElementType() reflect.Type {
	return reflect.TypeOf((*Todo)(nil))
}

func (i *Todo) ToTodoOutput() TodoOutput {
	return i.ToTodoOutputWithContext(context.Background())
}

func (i *Todo) ToTodoOutputWithContext(ctx context.Context) TodoOutput {
	return pulumi.ToOutputWithContext(ctx, i).(TodoOutput)
}

type TodoOutput struct {
	*pulumi.OutputState
}

func (TodoOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*Todo)(nil))
}

func (o TodoOutput) ToTodoOutput() TodoOutput {
	return o
}

func (o TodoOutput) ToTodoOutputWithContext(ctx context.Context) TodoOutput {
	return o
}

func init() {
	pulumi.RegisterOutputType(TodoOutput{})
}
