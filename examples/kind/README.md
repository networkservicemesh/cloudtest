# Kind shell provider example

CloudTest could be configured to control any kind of shell based clusters.

Running `cloudtest` inside this folder will trigger setup of Kind tool and execution 
of the few tests against this cloud, tests are dummy ones and just print current cluster config
passed. Some of these tests are failed to show how results will look.

## Configuration
Sample configuration contains 2 executions, one is go test folder and one is simple script.

Configuration could be checked here: [.cloudtest.yaml](./.cloudtest.yaml)

## Runtime information
During execution CloudTest will print execution statistics and information about cluster starting, etc.

```
INFO[0090] Statistics:
        Elapsed total: 1m30s
        Tests time: 1m30s
        Tasks  Completed: 0
               Remaining: 6 (~)

        Clusters:
                Cluster: kind Tasks left: 6
                        kind-1: starting, uptime: 1m31s

        Status  Passed: 0
        Status  Failed: 0
        Status  Timeout: 0
        Status  Skipped: 0 

```

## Results

### JUnit XML
After execution it will create `./results/junit.xml` file to integrate to CI systems.
Report file is extended with human readable comments and separated to suits based on executions defined. 
```xml
  <testsuites>
    <testsuite tests="8" failures="4" time="29.037307726" name="All tests">
        <properties></properties>
        <!--Suite was running for 29s-->
        <testsuite tests="1" failures="0" time="0.003621976" name="Sample shell tests">
            <properties></properties>
            <!--Suite was running for 0s-->
            <testsuite tests="1" failures="0" time="0.003621976" name="kind">
                <properties></properties>
                <!--Suite was running for 0s-->
                <testcase classname="" name="Sample shell tests" time="0.003621976" cluster_instance="kind-1"></testcase>
            </testsuite>
        </testsuite>
        <testsuite tests="7" failures="4" time="29.03368575" name="Sample go tests">
            <properties></properties>
            <!--Suite was running for 29s-->
            <testsuite tests="7" failures="4" time="29.03368575" name="kind">
                <properties></properties>
                <!--Suite was running for 29s-->
                <testcase classname="" name="TestTimeout" time="20.010373901" cluster_instance="kind-1">
                    <failure message="Test execution failed TestTimeout" type="ERROR">Execution attempt: 0 Output file: .results/kind-1/007-TestTimeout-run.log&#xA;Starting TestTimeout on kind-1&#xA;Command line go test . -test.timeout 20s -count 1 --run &#34;^(TestTimeout)$\\z&#34; --tags &#34;&#34; --test.v&#xA;env==[KUBECONFIG=/Users/user/Projects/NSM/cloudtest/examples/kind/.results/shell/kind-1/config ARTIFACTS_DIR=/Users/user/Projects/NSM/cloudtest/examples/kind/.results/kind-1/TestTimeout]&#xA;&#xA;=== RUN   TestTimeout&#xA;time=&#34;2021-02-16T11:56:35+07:00&#34; level=info msg=&#34;test timeout for 10 seconds:/Users/user/Projects/NSM/cloudtest/examples/kind/.results/shell/kind-1/config&#34;&#xA;TestTimeout: OnFail: running on fail script operations with KUBECONFIG=/Users/user/Projects/NSM/cloudtest/examples/kind/.results/shell/kind-1/config on cloud kind-1&#xA;Do cleanup on failure # In case of execution failure this script will cleanup after test</failure>
                </testcase>
                <testcase classname="" name="TestPass" time="1.655707858" cluster_instance="kind-1"></testcase>
                <testcase classname="" name="TestFail" time="1.071847017" cluster_instance="kind-1">
                    <failure message="Test execution failed TestFail" type="ERROR">Execution attempt: 0 Output file: .results/kind-1/010-TestFail-run.log&#xA;Starting TestFail on kind-1&#xA;Command line go test . -test.timeout 20s -count 1 --run &#34;^(TestFail)$\\z&#34; --tags &#34;&#34; --test.v&#xA;env==[KUBECONFIG=/Users/user/Projects/NSM/cloudtest/examples/kind/.results/shell/kind-1/config ARTIFACTS_DIR=/Users/user/Projects/NSM/cloudtest/examples/kind/.results/kind-1/TestFail]&#xA;&#xA;=== RUN   TestFail&#xA;time=&#34;2021-02-16T11:56:57+07:00&#34; level=info msg=&#34;Failed test: /Users/user/Projects/NSM/cloudtest/examples/kind/.results/shell/kind-1/config&#34;&#xA;--- FAIL: TestFail (0.00s)&#xA;FAIL&#xA;FAIL&#x9;github.com/networkservicemesh/cloudtest/examples/kind/tests&#x9;0.168s&#xA;FAIL&#xA;TestFail: OnFail: running on fail script operations with KUBECONFIG=/Users/user/Projects/NSM/cloudtest/examples/kind/.results/shell/kind-1/config on cloud kind-1&#xA;Do cleanup on failure # In case of execution failure this script will cleanup after test</failure>
                </testcase>
                <testsuite tests="2" failures="1" time="1.113919521" name="TestFailedSuite">
                    <properties></properties>
                    <testcase classname="" name="SetupSuite" time="0.001165" cluster_instance="kind-1">
                        <failure message="Test execution failed SetupSuite" type="ERROR">Execution attempt: 0 Output file: .results/kind-1/012-TestFailedSuite1-SetupSuite-run.log&#xA;=== RUN   TestFailedSuite&#xA;time=&#34;2021-02-16T11:56:56+07:00&#34; level=info msg=&#34;Failed suite: /Users/user/Projects/NSM/cloudtest/examples/kind/.results/shell/kind-1/config&#34;&#xA;--- FAIL: TestFailedSuite (0.00s)</failure>
                    </testcase>
                    <testcase classname="" name="TestPass" time="0" cluster_instance="kind-1">
                        <skipped message="Suite setup failed"></skipped>
                    </testcase>
                    <testcase classname="" name="TestFail" time="0" cluster_instance="kind-1">
                        <skipped message="Suite setup failed"></skipped>
                    </testcase>
                </testsuite>
                <testsuite tests="2" failures="1" time="5.181837453" name="TestSuite">
                    <properties></properties>
                    <testcase classname="" name="TestPass" time="0.00039" cluster_instance="kind-1"></testcase>
                    <testcase classname="" name="TestFail" time="0.000406" cluster_instance="kind-1">
                        <failure message="Test execution failed TestFail" type="ERROR">Execution attempt: 0 Output file: .results/kind-1/016-TestSuite1-TestFail-run.log&#xA;=== RUN   TestSuite/TestFail&#xA;time=&#34;2021-02-16T11:56:31+07:00&#34; level=info msg=&#34;Failed test: /Users/user/Projects/NSM/cloudtest/examples/kind/.results/shell/kind-1/config&#34;&#xA;--- FAIL: TestSuite/TestFail (0.00s)</failure>
                    </testcase>
                </testsuite>
            </testsuite>
        </testsuite>
    </testsuite>
</testsuites>
```

### Logs structure

During execution for every operation CloudTest produce log files, start of cluster instances, 
execution of tests all are marked as operation and be stored into separate files.

![Structure](./images/CloudTest_kind_results.png) 

All operations could be checked in case of test failures or cluster instance errors.