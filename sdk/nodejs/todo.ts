// *** WARNING: this file was generated by the Pulumi SDK Generator. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

import * as pulumi from "@pulumi/pulumi";
import * as utilities from "./utilities";

export class Todo extends pulumi.CustomResource {
    /**
     * Get an existing Todo resource's state with the given name, ID, and optional extra
     * properties used to qualify the lookup.
     *
     * @param name The _unique_ name of the resulting resource.
     * @param id The _unique_ provider ID of the resource to lookup.
     * @param opts Optional settings to control the behavior of the CustomResource.
     */
    public static get(name: string, id: pulumi.Input<pulumi.ID>, opts?: pulumi.CustomResourceOptions): Todo {
        return new Todo(name, undefined as any, { ...opts, id: id });
    }

    /** @internal */
    public static readonly __pulumiType = 'xyz:index:Todo';

    /**
     * Returns true if the given object is an instance of Todo.  This is designed to work even
     * when multiple copies of the Pulumi SDK have been loaded into the same process.
     */
    public static isInstance(obj: any): obj is Todo {
        if (obj === undefined || obj === null) {
            return false;
        }
        return obj['__pulumiType'] === Todo.__pulumiType;
    }

    public readonly completed!: pulumi.Output<boolean>;
    public readonly order!: pulumi.Output<number>;
    public readonly title!: pulumi.Output<string>;
    public readonly url!: pulumi.Output<string>;

    /**
     * Create a Todo resource with the given unique name, arguments, and options.
     *
     * @param name The _unique_ name of the resource.
     * @param args The arguments to use to populate this resource's properties.
     * @param opts A bag of options that control this resource's behavior.
     */
    constructor(name: string, args: TodoArgs, opts?: pulumi.CustomResourceOptions) {
        let inputs: pulumi.Inputs = {};
        opts = opts || {};
        if (!opts.id) {
            if ((!args || args.title === undefined) && !opts.urn) {
                throw new Error("Missing required property 'title'");
            }
            inputs["completed"] = args ? args.completed : undefined;
            inputs["order"] = args ? args.order : undefined;
            inputs["title"] = args ? args.title : undefined;
            inputs["url"] = args ? args.url : undefined;
        } else {
            inputs["completed"] = undefined /*out*/;
            inputs["order"] = undefined /*out*/;
            inputs["title"] = undefined /*out*/;
            inputs["url"] = undefined /*out*/;
        }
        if (!opts.version) {
            opts = pulumi.mergeOptions(opts, { version: utilities.getVersion()});
        }
        super(Todo.__pulumiType, name, inputs, opts);
    }
}

/**
 * The set of arguments for constructing a Todo resource.
 */
export interface TodoArgs {
    readonly completed?: pulumi.Input<boolean>;
    readonly order?: pulumi.Input<number>;
    readonly title: pulumi.Input<string>;
    readonly url?: pulumi.Input<string>;
}
