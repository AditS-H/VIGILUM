// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.28;

import "forge-std/Script.sol";
import "../src/RedTeamDAO.sol";
import "../src/ProofOfExploit.sol";
import "../src/VigilumRegistry.sol";

contract DeployDAOContracts is Script {
    function setUp() public {}

    function run() external {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        address vigilumRegistry = vm.envAddress("VIGILUM_REGISTRY");

        vm.startBroadcast(deployerPrivateKey);

        // Deploy RedTeamDAO
        RedTeamDAO dao = new RedTeamDAO();
        console.log("RedTeamDAO deployed to:", address(dao));

        // Deploy ProofOfExploit
        ProofOfExploit proofOfExploit = new ProofOfExploit(
            vigilumRegistry,
            address(dao)
        );
        console.log("ProofOfExploit deployed to:", address(proofOfExploit));

        // Set ProofOfExploit address in DAO
        // (would need setter method in RedTeamDAO)

        vm.stopBroadcast();

        console.log("\n=== Deployment Summary ===");
        console.log("RedTeamDAO:", address(dao));
        console.log("ProofOfExploit:", address(proofOfExploit));
        console.log("VigilumRegistry:", vigilumRegistry);
    }
}
