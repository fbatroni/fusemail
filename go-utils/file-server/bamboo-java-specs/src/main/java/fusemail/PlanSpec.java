package fusemail;

import com.atlassian.bamboo.specs.api.BambooSpec;
import com.atlassian.bamboo.specs.api.builders.credentials.SharedCredentialsIdentifier;
import com.atlassian.bamboo.specs.api.builders.notification.Notification;
import com.atlassian.bamboo.specs.api.builders.permission.PermissionType;
import com.atlassian.bamboo.specs.api.builders.permission.Permissions;
import com.atlassian.bamboo.specs.api.builders.permission.PlanPermissions;
import com.atlassian.bamboo.specs.api.builders.plan.Job;
import com.atlassian.bamboo.specs.api.builders.plan.Plan;
import com.atlassian.bamboo.specs.api.builders.plan.PlanIdentifier;
import com.atlassian.bamboo.specs.api.builders.plan.Stage;
import com.atlassian.bamboo.specs.api.builders.plan.branches.BranchCleanup;
import com.atlassian.bamboo.specs.api.builders.plan.branches.PlanBranchManagement;
import com.atlassian.bamboo.specs.api.builders.project.Project;
import com.atlassian.bamboo.specs.api.builders.repository.VcsChangeDetection;
import com.atlassian.bamboo.specs.builders.notification.CommittersRecipient;
import com.atlassian.bamboo.specs.builders.notification.PlanStatusChangedNotification;
import com.atlassian.bamboo.specs.builders.repository.git.GitRepository;
import com.atlassian.bamboo.specs.builders.task.CheckoutItem;
import com.atlassian.bamboo.specs.builders.task.ScriptTask;
import com.atlassian.bamboo.specs.builders.task.TestParserTask;
import com.atlassian.bamboo.specs.builders.task.VcsCheckoutTask;
import com.atlassian.bamboo.specs.builders.trigger.RemoteTrigger;
import com.atlassian.bamboo.specs.model.task.TestParserTaskProperties;
import com.atlassian.bamboo.specs.util.BambooServer;

/**
 * Plan configuration for Bamboo.
 * Learn more on: <a href="https://confluence.atlassian.com/display/BAMBOO/Bamboo+Specs">https://confluence.atlassian.com/display/BAMBOO/Bamboo+Specs</a>
 */
@BambooSpec
public class PlanSpec {
    public static final String BAMBOO_SERVER = "http://bamboo2203a.electric.net:8085";
    public static final String PROJECT_NAME = "Templates";
    public static final String PROJECT_KEY = "TMPLTS";
    public static final String REPO_NAME = "fm-app-go-template";
    public static final String PLAN_NAME = REPO_NAME;
    public static final String PLAN_KEY = "FAGT";

    /**
     * Run main to publish plan on Bamboo
     */
    public static void main(final String[] args) throws Exception {
        //By default credentials are read from the '.credentials' file.
        BambooServer bambooServer = new BambooServer(BAMBOO_SERVER);
        final PlanSpec planSpec = new PlanSpec();

        final Plan plan = planSpec.createPlan();
        bambooServer.publish(plan);

        final PlanPermissions planPermission = planSpec.planPermission();
        bambooServer.publish(planPermission);
    }

    Project project() {
        return new Project()
                .name(PROJECT_NAME)
                .key(PROJECT_KEY);
    }

    Plan createPlan() {
        return new Plan(project(), PLAN_NAME, PLAN_KEY)
                .description("This plan is being managed via configuration-as-code. Modify " + REPO_NAME + " project to update the plan.")
                .stages(new Stage("Default Stage")
                        .description("Build, test and push")
                        .jobs(new Job("Default Job", "DJ")
                                .tasks(new VcsCheckoutTask()
                                                .description("Checkout " + REPO_NAME + " repo")
                                                .checkoutItems(new CheckoutItem().defaultRepository())
                                                .cleanCheckout(true),
                                        new ScriptTask()
                                                .description("Fix version in Makefile for bamboo env")
                                                .inlineBody("sed -i -e \"s/^VERSION=.*/VERSION=$(git describe --tags --always)/g\" Makefile"),
                                        new ScriptTask()
                                                .description("Build Docker image")
                                                .inlineBody("make docker"),
                                        new ScriptTask()
                                                .description("Run test")
                                                .inlineBody("make docker-test"),
                                        new ScriptTask()
                                                .description("Push to Docker registry")
                                                .inlineBody("make docker-push"),
                                        new ScriptTask()
                                                .description("Push to artifact server")
                                                .inlineBody("make deploy"))
                                .finalTasks(new TestParserTask(TestParserTaskProperties.TestType.JUNIT)
                                                .description("Parse test result")
                                                .resultDirectories("test-result/junit-test-report.xml"),
                                        new ScriptTask()
                                                .description("Cleanup build and test results")
                                                .inlineBody("make clean"))))
                .planRepositories(new GitRepository()
                        .name(REPO_NAME)
                        .url("git@bitbucket.org:fusemail/" + REPO_NAME + ".git")
                        .branch("master")
                        .authentication(new SharedCredentialsIdentifier("Bitbucket Cloud - Bamboo Master Key"))
                        .changeDetection(new VcsChangeDetection()))
                .notifications(new Notification()
                        .recipients(new CommittersRecipient())
                        .type(new PlanStatusChangedNotification()))
                .triggers(new RemoteTrigger()
                        .name("Remote trigger")
                        .description("Bitbucket Cloud Trigger")
                        .triggerIPAddresses("104.192.136.0/21,34.198.203.127,34.198.178.64,34.198.32.85,104.192.142.195,10.105.6.166"))
                .planBranchManagement(new PlanBranchManagement()
                        .createForVcsBranch()
                        .delete(new BranchCleanup()
                                .whenRemovedFromRepositoryAfterDays(0))
                        .notificationForCommitters());
    }

    public PlanPermissions planPermission() {
        final PlanPermissions planPermission = new PlanPermissions(new PlanIdentifier(PROJECT_KEY, PLAN_KEY))
                .permissions(new Permissions()
                        .loggedInUserPermissions(PermissionType.VIEW)
                        .anonymousUserPermissionView());
        return planPermission;
    }
}
