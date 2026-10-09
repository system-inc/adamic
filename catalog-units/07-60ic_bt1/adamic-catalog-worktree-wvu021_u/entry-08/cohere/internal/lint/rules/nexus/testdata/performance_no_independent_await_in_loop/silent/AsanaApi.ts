// Shape from ahra modules/asana/AsanaApi.ts:126 (listTasks).
// A for-of over sections that prints a header, awaits the section's tasks, prints them, and prints
// a blank line. The output order is the point, so the loop is sequential on purpose. The inner
// for-of holds no await.
interface AsanaTaskInterface {
    name: string;
}
interface AsanaResponseInterface {
    data?: Array<{ gid: string; name: string }>;
}
declare const console: { log(...values: unknown[]): void };

export class AsanaApi {
    static async request(path: string): Promise<AsanaResponseInterface | null> {
        return { data: [{ gid: path, name: path }] };
    }
    static sortTasks(tasks: AsanaTaskInterface[]): AsanaTaskInterface[] {
        return tasks;
    }
    static formatTaskLine(task: AsanaTaskInterface, indent = ''): string {
        return indent + task.name;
    }

    static async listTasks(projectGid: string, showCompleted: boolean = false): Promise<void> {
        const fields = 'name,completed,completed_at,modified_at,due_on,assignee.name';

        const sections = await this.request(`/projects/${projectGid}/sections?opt_fields=name`);

        if(sections?.data?.length) {
            console.log('=== TASKS ===\n');
            for(const section of sections.data) {
                console.log(`\u2500\u2500 ${section.name} \u2500\u2500`);
                const completedFilter = showCompleted ? '' : '&completed_since=now';
                const sectionTasks = await this.request(
                    `/sections/${section.gid}/tasks?opt_fields=${fields}&limit=100${completedFilter}`,
                );
                if(sectionTasks?.data?.length) {
                    for(const task of this.sortTasks(sectionTasks.data as unknown as AsanaTaskInterface[])) {
                        console.log(this.formatTaskLine(task, '  '));
                    }
                }
                console.log('');
            }
        }
    }
}
