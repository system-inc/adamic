export declare enum ProgramUpdateLevel {
    /** Program is updated with same root file names and options */
    Update = 0,
    /** Loads program after updating root file names from the disk */
    RootNamesAndUpdate = 1,
    /**
     * Loads program completely, including:
     *  - re-reading contents of config file from disk
     *  - calculating root file names for the program
     *  - Updating the program
     */
    Full = 2
}
