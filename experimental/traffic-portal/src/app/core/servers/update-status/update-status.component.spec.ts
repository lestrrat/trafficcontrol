/*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/
import { HttpClientModule } from "@angular/common/http";
import { ComponentFixture, TestBed } from "@angular/core/testing";

import {ServerService} from "../../../shared/api";
import { UpdateStatusComponent } from "./update-status.component";

describe("UpdateStatusComponent", () => {
	let component: UpdateStatusComponent;
	let fixture: ComponentFixture<UpdateStatusComponent>;

	beforeEach(async () => {
		const mockAPIService = jasmine.createSpyObj(["getServers", "getStatuses"]);
		mockAPIService.getStatuses.and.returnValue(new Promise(r => r([])));
		await TestBed.configureTestingModule({
			declarations: [ UpdateStatusComponent ],
			imports: [ HttpClientModule ],
			providers: [
				{ provide: ServerService, useValue: mockAPIService }
			]
		})
			.compileComponents();
	});

	beforeEach(() => {
		fixture = TestBed.createComponent(UpdateStatusComponent);
		component = fixture.componentInstance;
		fixture.detectChanges();
	});

	it("should create", () => {
		expect(component).toBeTruthy();
	});
});
