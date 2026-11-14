import { assertType, createQueue, Debug, JsTyping, } from "./_namespaces/ts.js";
import { ActionInvalidate, ActionPackageInstalled, ActionSet, ActionWatchTypingLocations, createInstallTypingsRequest, EventBeginInstallTypes, EventEndInstallTypes, EventInitializationFailed, EventTypesRegistry, LogLevel, stringifyIndented, } from "./_namespaces/ts.server.js";
/** @internal */
export class TypingsInstallerAdapter {
    telemetryEnabled;
    logger;
    host;
    globalTypingsCacheLocation;
    event;
    maxActiveRequestCount;
    installer;
    projectService;
    activeRequestCount = 0;
    requestQueue = createQueue();
    requestMap = new Map(); // Maps project name to newest requestQueue entry for that project
    /** We will lazily request the types registry on the first call to `isKnownTypesPackageName` and store it in `typesRegistryCache`. */
    requestedRegistry = false;
    typesRegistryCache;
    // This number is essentially arbitrary.  Processing more than one typings request
    // at a time makes sense, but having too many in the pipe results in a hang
    // (see https://github.com/nodejs/node/issues/7657).
    // It would be preferable to base our limit on the amount of space left in the
    // buffer, but we have yet to find a way to retrieve that value.
    static requestDelayMillis = 100;
    packageInstalledPromise;
    packageInstallId = 0;
    constructor(telemetryEnabled, logger, host, globalTypingsCacheLocation, event, maxActiveRequestCount) {
        this.telemetryEnabled = telemetryEnabled;
        this.logger = logger;
        this.host = host;
        this.globalTypingsCacheLocation = globalTypingsCacheLocation;
        this.event = event;
        this.maxActiveRequestCount = maxActiveRequestCount;
    }
    isKnownTypesPackageName(name) {
        // We want to avoid looking this up in the registry as that is expensive. So first check that it's actually an NPM package.
        const validationResult = JsTyping.validatePackageName(name);
        if (validationResult !== JsTyping.NameValidationResult.Ok) {
            return false;
        }
        if (!this.requestedRegistry) {
            this.requestedRegistry = true;
            this.installer.send({ kind: "typesRegistry" });
        }
        return !!this.typesRegistryCache?.has(name);
    }
    installPackage(options) {
        this.packageInstallId++;
        const request = { kind: "installPackage", ...options, id: this.packageInstallId };
        const promise = new Promise((resolve, reject) => {
            (this.packageInstalledPromise ??= new Map()).set(this.packageInstallId, { resolve, reject });
        });
        this.installer.send(request);
        return promise;
    }
    attach(projectService) {
        this.projectService = projectService;
        this.installer = this.createInstallerProcess();
    }
    onProjectClosed(p) {
        this.installer.send({ projectName: p.getProjectName(), kind: "closeProject" });
    }
    enqueueInstallTypingsRequest(project, typeAcquisition, unresolvedImports) {
        const request = createInstallTypingsRequest(project, typeAcquisition, unresolvedImports);
        if (this.logger.hasLevel(LogLevel.verbose)) {
            this.logger.info(`TIAdapter:: Scheduling throttled operation:${stringifyIndented(request)}`);
        }
        if (this.activeRequestCount < this.maxActiveRequestCount) {
            this.scheduleRequest(request);
        }
        else {
            if (this.logger.hasLevel(LogLevel.verbose)) {
                this.logger.info(`TIAdapter:: Deferring request for: ${request.projectName}`);
            }
            this.requestQueue.enqueue(request);
            this.requestMap.set(request.projectName, request);
        }
    }
    handleMessage(response) {
        if (this.logger.hasLevel(LogLevel.verbose)) {
            this.logger.info(`TIAdapter:: Received response:${stringifyIndented(response)}`);
        }
        switch (response.kind) {
            case EventTypesRegistry:
                this.typesRegistryCache = new Map(Object.entries(response.typesRegistry));
                break;
            case ActionPackageInstalled: {
                const promise = this.packageInstalledPromise?.get(response.id);
                Debug.assertIsDefined(promise, "Should find the promise for package install");
                this.packageInstalledPromise?.delete(response.id);
                if (response.success) {
                    promise.resolve({ successMessage: response.message });
                }
                else {
                    promise.reject(response.message);
                }
                this.projectService.updateTypingsForProject(response);
                // The behavior is the same as for setTypings, so send the same event.
                this.event(response, "setTypings");
                break;
            }
            case EventInitializationFailed: {
                const body = {
                    message: response.message,
                };
                const eventName = "typesInstallerInitializationFailed";
                this.event(body, eventName);
                break;
            }
            case EventBeginInstallTypes: {
                const body = {
                    eventId: response.eventId,
                    packages: response.packagesToInstall,
                };
                const eventName = "beginInstallTypes";
                this.event(body, eventName);
                break;
            }
            case EventEndInstallTypes: {
                if (this.telemetryEnabled) {
                    const body = {
                        telemetryEventName: "typingsInstalled",
                        payload: {
                            installedPackages: response.packagesToInstall.join(","),
                            installSuccess: response.installSuccess,
                            typingsInstallerVersion: response.typingsInstallerVersion,
                        },
                    };
                    const eventName = "telemetry";
                    this.event(body, eventName);
                }
                const body = {
                    eventId: response.eventId,
                    packages: response.packagesToInstall,
                    success: response.installSuccess,
                };
                const eventName = "endInstallTypes";
                this.event(body, eventName);
                break;
            }
            case ActionInvalidate: {
                this.projectService.updateTypingsForProject(response);
                break;
            }
            case ActionSet: {
                if (this.activeRequestCount > 0) {
                    this.activeRequestCount--;
                }
                else {
                    Debug.fail("TIAdapter:: Received too many responses");
                }
                while (!this.requestQueue.isEmpty()) {
                    const queuedRequest = this.requestQueue.dequeue();
                    if (this.requestMap.get(queuedRequest.projectName) === queuedRequest) {
                        this.requestMap.delete(queuedRequest.projectName);
                        this.scheduleRequest(queuedRequest);
                        break;
                    }
                    if (this.logger.hasLevel(LogLevel.verbose)) {
                        this.logger.info(`TIAdapter:: Skipping defunct request for: ${queuedRequest.projectName}`);
                    }
                }
                this.projectService.updateTypingsForProject(response);
                this.event(response, "setTypings");
                break;
            }
            case ActionWatchTypingLocations:
                this.projectService.watchTypingLocations(response);
                break;
            default:
                assertType(response);
        }
    }
    scheduleRequest(request) {
        if (this.logger.hasLevel(LogLevel.verbose)) {
            this.logger.info(`TIAdapter:: Scheduling request for: ${request.projectName}`);
        }
        this.activeRequestCount++;
        (() => {
            const timerHost = this.host;
            const callback = (_label) => {
                if (this.logger.hasLevel(LogLevel.verbose)) {
                    this.logger.info(`TIAdapter:: Sending request:${stringifyIndented(request)}`);
                }
                this.installer.send(request);
            };
            const timeoutDelay = TypingsInstallerAdapter.requestDelayMillis;
            const t1 = `${request.projectName}::${request.kind}`;
            return timerHost.setTimeout(() => callback(t1), timeoutDelay, t1);
        })();
    }
}
